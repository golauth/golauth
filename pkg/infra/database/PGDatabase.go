package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/subosito/gotenv"
)

type PGDatabase struct {
	db *sql.DB
}

func NewPGDatabase() Database {
	_ = gotenv.Load()
	s := loadDBSettings()

	db, err := openPool(s)
	if err != nil {
		logrus.Fatal(err)
	}
	// A wedged or unreachable database now fails the boot within pingTimeout
	// instead of hanging it indefinitely on a bare db.Ping().
	if err := verifyConnection(context.Background(), db, s.pingTimeout); err != nil {
		_ = db.Close()
		logrus.Fatal(err)
	}

	pg := &PGDatabase{db: db}
	if s.runMigrations {
		if err := pg.migrate(s.migrationSourceURL); err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Warn("RUN_MIGRATIONS is false: skipping migrations at boot (run them as a separate job)")
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
	logrus.Info("starting migration execution")
	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("database: could not create migration connection: %w", err)
	}
	logrus.Infof("Executing migrations on path: %s", sourceURL)
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
	logrus.Info("finalizing migrations!")
	return nil
}

func (d PGDatabase) Close() {
	err := d.db.Close()
	if err != nil {
		logrus.Error(err)
	}
}
