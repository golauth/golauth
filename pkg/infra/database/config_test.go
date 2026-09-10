package database

import (
	"context"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestDSNQuotesEveryValue(t *testing.T) {
	base := dbSettings{host: "db", port: "5432", name: "golauth", user: "svc", sslmode: "require"}

	cases := []struct {
		name     string
		password string
		want     string
	}{
		{
			name:     "plain",
			password: "s3cr3t",
			want:     "host='db' port='5432' dbname='golauth' user='svc' password='s3cr3t' sslmode='require'",
		},
		{
			name:     "space",
			password: "pa ss word",
			want:     "host='db' port='5432' dbname='golauth' user='svc' password='pa ss word' sslmode='require'",
		},
		{
			name:     "single quote",
			password: "pa'ss",
			want:     `host='db' port='5432' dbname='golauth' user='svc' password='pa\'ss' sslmode='require'`,
		},
		{
			name:     "backslash",
			password: `pa\ss`,
			want:     `host='db' port='5432' dbname='golauth' user='svc' password='pa\\ss' sslmode='require'`,
		},
		{
			name:     "quote and backslash and space",
			password: `p' a\ b`,
			want:     `host='db' port='5432' dbname='golauth' user='svc' password='p\' a\\ b' sslmode='require'`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := base
			s.password = c.password
			got := dsn(s)
			require.Equal(t, c.want, got)
			// lib/pq must accept and round-trip the string we built.
			_, err := pq.NewConnector(got)
			require.NoError(t, err, "lib/pq rejected the DSN: %s", got)
		})
	}
}

func TestDSNIncludesSSLRootCertOnlyWhenSet(t *testing.T) {
	s := dbSettings{host: "db", port: "5432", name: "n", user: "u", password: "p", sslmode: "verify-full"}
	require.NotContains(t, dsn(s), "sslrootcert")

	s.sslrootcert = "/etc/ssl/certs/rds.pem"
	require.Contains(t, dsn(s), `sslrootcert='/etc/ssl/certs/rds.pem'`)
}

func TestLoadDBSettingsDefaults(t *testing.T) {
	for _, k := range []string{
		"DB_SSLMODE", "DB_SSLROOTCERT", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_IDLE_TIME", "DB_PING_TIMEOUT", "RUN_MIGRATIONS",
	} {
		t.Setenv(k, "")
	}

	s := loadDBSettings()
	require.Equal(t, "disable", s.sslmode)
	require.Empty(t, s.sslrootcert)
	require.Equal(t, 25, s.maxOpenConns)
	require.Equal(t, 25, s.maxIdleConns)
	require.Equal(t, 30*time.Minute, s.connMaxLifetime)
	require.Equal(t, 5*time.Minute, s.connMaxIdleTime)
	require.Equal(t, 5*time.Second, s.pingTimeout)
	require.True(t, s.runMigrations)
}

func TestLoadDBSettingsOverrides(t *testing.T) {
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("DB_SSLROOTCERT", "/ca.pem")
	t.Setenv("DB_MAX_OPEN_CONNS", "7")
	t.Setenv("DB_MAX_IDLE_CONNS", "3")
	t.Setenv("DB_CONN_MAX_LIFETIME", "10m")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "90s")
	t.Setenv("DB_PING_TIMEOUT", "2s")
	t.Setenv("RUN_MIGRATIONS", "false")

	s := loadDBSettings()
	require.Equal(t, "require", s.sslmode)
	require.Equal(t, "/ca.pem", s.sslrootcert)
	require.Equal(t, 7, s.maxOpenConns)
	require.Equal(t, 3, s.maxIdleConns)
	require.Equal(t, 10*time.Minute, s.connMaxLifetime)
	require.Equal(t, 90*time.Second, s.connMaxIdleTime)
	require.Equal(t, 2*time.Second, s.pingTimeout)
	require.False(t, s.runMigrations)
}

func TestLoadDBSettingsInvalidValuesFallBackToDefault(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "-4")
	t.Setenv("DB_CONN_MAX_LIFETIME", "not-a-duration")
	t.Setenv("RUN_MIGRATIONS", "maybe")

	s := loadDBSettings()
	require.Equal(t, 25, s.maxOpenConns)
	require.Equal(t, 30*time.Minute, s.connMaxLifetime)
	require.True(t, s.runMigrations)
}

func TestOpenPoolAppliesBoundedPool(t *testing.T) {
	s := dbSettings{
		host: "db", port: "5432", name: "n", user: "u", password: "p", sslmode: "disable",
		maxOpenConns: 7, maxIdleConns: 3,
		connMaxLifetime: time.Minute, connMaxIdleTime: time.Minute,
	}
	db, err := openPool(s)
	require.NoError(t, err) // sql.Open does not connect
	t.Cleanup(func() { _ = db.Close() })

	require.Equal(t, 7, db.Stats().MaxOpenConnections)
}

// A boot against an unroutable host must fail within the ping timeout, not hang.
func TestVerifyConnectionFailsFastOnUnreachableHost(t *testing.T) {
	s := dbSettings{
		host: "192.0.2.1", port: "5432", name: "n", user: "u", password: "p", sslmode: "disable",
		maxOpenConns: 1, maxIdleConns: 1,
	}
	db, err := openPool(s)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	start := time.Now()
	err = verifyConnection(context.Background(), db, time.Second)
	elapsed := time.Since(start)

	require.Error(t, err)
	require.Less(t, elapsed, 5*time.Second, "ping did not honour the timeout")
}
