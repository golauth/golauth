package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/subosito/gotenv"
)

// fatal logs a boot-time failure and exits non-zero. The database handle is
// built during process start, before there is anything to gracefully unwind, so
// a hard exit is the honest behaviour; slog has no Fatal of its own.
func fatal(msg string, err error) {
	slog.Error(msg, "err", err.Error())
	os.Exit(1)
}

type PGDatabase struct {
	db *sql.DB
}

func NewPGDatabase() Database {
	_ = gotenv.Load()
	s := loadDBSettings()

	db, err := openPool(s)
	if err != nil {
		fatal("database: could not open connection pool", err)
	}
	// A wedged or unreachable database now fails the boot within pingTimeout
	// instead of hanging it indefinitely on a bare db.Ping().
	if err := verifyConnection(context.Background(), db, s.pingTimeout); err != nil {
		_ = db.Close()
		fatal("database: connection check failed", err)
	}

	pg := &PGDatabase{db: db}
	if s.runMigrations {
		if err := pg.migrate(s.migrationSourceURL); err != nil {
			fatal("database: migration failed", err)
		}
	} else {
		slog.Warn("RUN_MIGRATIONS is false: skipping migrations at boot (run them as a separate job)")
	}
	return pg
}

func (d PGDatabase) Many(ctx context.Context, query string, params ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, params...)
}

func (d PGDatabase) One(ctx context.Context, query string, params ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, params...)
}

func (d PGDatabase) Exec(ctx context.Context, query string, params ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, params...)
}

// Ping verifies the connection pool can reach Postgres, honouring ctx's
// deadline (pq.Connector does; the string DSN form would not).
func (d PGDatabase) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// openPool builds the driver handle and applies the bounded pool configuration.
// It goes through pq.NewConnector rather than sql.Open("postgres", dsn) so that
// PingContext (and every later query) honours its context deadline: the string
// form of the driver ignores the context during the initial dial.
func openPool(s dbSettings) (*sql.DB, error) {
	connector, err := pq.NewConnector(dsn(s))
	if err != nil {
		return nil, fmt.Errorf("database: could not parse connection settings: %w", err)
	}
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(s.maxOpenConns)
	db.SetMaxIdleConns(s.maxIdleConns)
	db.SetConnMaxLifetime(s.connMaxLifetime)
	db.SetConnMaxIdleTime(s.connMaxIdleTime)
	return db, nil
}

func verifyConnection(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database: could not establish connection within %s: %w", timeout, err)
	}
	return nil
}

func (d PGDatabase) migrate(sourceURL string) error {
	slog.Info("starting migration execution")
	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("database: could not create migration connection: %w", err)
	}
	slog.Info("executing migrations", "source", sourceURL)
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+sourceURL,
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("database: could not prepare database migration: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database: error when executing database migration: %w", err)
	}
	slog.Info("finalizing migrations")
	return nil
}

func (d PGDatabase) Close() {
	if err := d.db.Close(); err != nil {
		slog.Error("database: error closing connection pool", "err", err.Error())
	}
}
