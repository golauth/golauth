package tests

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// appTables are the tables every migration set is expected to create. It
// excludes migrate's own schema_migrations bookkeeping table.
var appTables = []string{
	"golauth_authority",
	"golauth_login_attempt",
	"golauth_refresh_token",
	"golauth_role",
	"golauth_role_authority",
	"golauth_user",
	"golauth_user_role",
}

// bcryptHash matches a bcrypt digest in any of its prefixes. A migration that
// contains one is seeding a credential.
var bcryptHash = regexp.MustCompile(`\$2[aby]\$[0-9]{2}\$`)

// TestNoMigrationSeedsACredential is a crude but effective regression guard:
// no migration may create a user or embed a password hash. The shipped default
// admin (its hash was in version control, so every untouched install was
// compromised) is exactly the thing this stops from coming back. Bootstrap the
// first admin from BOOTSTRAP_ADMIN_* instead.
func TestNoMigrationSeedsACredential(t *testing.T) {
	files, err := filepath.Glob("../migrations/*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, files, "no migration files found")

	insertUser := regexp.MustCompile(`(?i)insert\s+into\s+golauth_user\b`)
	checked := 0
	for _, f := range files {
		if strings.HasSuffix(f, ".down.sql") {
			continue // a down file deletes rows; that is the point of it
		}
		body, err := os.ReadFile(f)
		require.NoError(t, err)
		require.False(t, insertUser.Match(body), "%s inserts into golauth_user", f)
		require.False(t, bcryptHash.Match(body), "%s embeds a password hash", f)
		checked++
	}
	require.Positive(t, checked, "no up migrations were scanned")
}

// TestMigrationsFullDownUpCycle proves every migration has a working down path:
// apply all, roll all back, apply all again, against a real container.
func TestMigrationsFullDownUpCycle(t *testing.T) {
	ctx, err := ContainerDBStart("./..")
	require.NoError(t, err)
	defer ContainerDBStop(ctx)

	db, err := sql.Open("postgres", dsnFromEnv())
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	require.NoError(t, db.Ping())

	m := newMigrator(t, db)
	defer func() { _, _ = m.Close() }()

	require.NoError(t, m.Up(), "initial up")
	require.ElementsMatch(t, appTables, listAppTables(t, db), "tables after first up")

	require.NoError(t, m.Down(), "full rollback")
	require.Empty(t, listAppTables(t, db), "every golauth_* table must be gone after down")

	require.NoError(t, m.Up(), "re-apply after rollback")
	require.ElementsMatch(t, appTables, listAppTables(t, db), "tables after re-apply")
}

func newMigrator(t *testing.T, db *sql.DB) *migrate.Migrate {
	t.Helper()
	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	require.NoError(t, err)
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+os.Getenv("MIGRATION_SOURCE_URL"), "postgres", driver)
	require.NoError(t, err)
	return m
}

func dsnFromEnv() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
}

func listAppTables(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query(`
		SELECT table_name FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name LIKE 'golauth_%'`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	var got []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		got = append(got, name)
	}
	require.NoError(t, rows.Err())
	sort.Strings(got)
	return got
}
