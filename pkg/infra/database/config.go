package database

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Defaults for the connection pool and boot behaviour. Idle equal to open
// avoids reconnect churn under steady load; the lifetime caps bound how long a
// single TCP connection (and its server-side backend) lives.
const (
	defaultSSLMode         = "disable"
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 25
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
	defaultPingTimeout     = 5 * time.Second
	defaultRunMigrations   = true
)

// dbSettings is everything NewPGDatabase reads from the environment.
type dbSettings struct {
	host, port, name, user, password string
	sslmode, sslrootcert             string
	maxOpenConns, maxIdleConns       int
	connMaxLifetime, connMaxIdleTime time.Duration
	pingTimeout                      time.Duration
	runMigrations                    bool
	migrationSourceURL               string
}

// loadDBSettings reads the DB_* environment. Every knob has a safe default so an
// unset environment still boots against a local Postgres.
func loadDBSettings() dbSettings {
	return dbSettings{
		host:               os.Getenv("DB_HOST"),
		port:               os.Getenv("DB_PORT"),
		name:               os.Getenv("DB_NAME"),
		user:               os.Getenv("DB_USERNAME"),
		password:           os.Getenv("DB_PASSWORD"),
		sslmode:            envStr("DB_SSLMODE", defaultSSLMode),
		sslrootcert:        os.Getenv("DB_SSLROOTCERT"),
		maxOpenConns:       envInt("DB_MAX_OPEN_CONNS", defaultMaxOpenConns),
		maxIdleConns:       envInt("DB_MAX_IDLE_CONNS", defaultMaxIdleConns),
		connMaxLifetime:    envDuration("DB_CONN_MAX_LIFETIME", defaultConnMaxLifetime),
		connMaxIdleTime:    envDuration("DB_CONN_MAX_IDLE_TIME", defaultConnMaxIdleTime),
		pingTimeout:        envDuration("DB_PING_TIMEOUT", defaultPingTimeout),
		runMigrations:      envBool("RUN_MIGRATIONS", defaultRunMigrations),
		migrationSourceURL: os.Getenv("MIGRATION_SOURCE_URL"),
	}
}

// dsn builds a lib/pq keyword/value connection string. Every value is single
// quoted with backslash and quote escaped, so a password containing a space, a
// single quote or a backslash is transmitted intact -- fmt.Sprintf
// concatenation mishandled all three.
func dsn(s dbSettings) string {
	parts := []string{
		"host=" + quoteDSNValue(s.host),
		"port=" + quoteDSNValue(s.port),
		"dbname=" + quoteDSNValue(s.name),
		"user=" + quoteDSNValue(s.user),
		"password=" + quoteDSNValue(s.password),
		"sslmode=" + quoteDSNValue(s.sslmode),
	}
	if s.sslrootcert != "" {
		parts = append(parts, "sslrootcert="+quoteDSNValue(s.sslrootcert))
	}
	return strings.Join(parts, " ")
}

var dsnValueEscaper = strings.NewReplacer(`\`, `\\`, `'`, `\'`)

func quoteDSNValue(v string) string {
	return "'" + dsnValueEscaper.Replace(v) + "'"
}

func envStr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
		slog.Warn("invalid env var, using default", "key", key, "value", os.Getenv(key), "default", def)
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		slog.Warn("invalid duration env var, using default",
			"key", key, "value", os.Getenv(key), "want", "a Go duration like \"30m\"", "default", def.String())
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
		slog.Warn("invalid boolean env var, using default",
			"key", key, "value", os.Getenv(key), "want", "true or false", "default", def)
	}
	return def
}
