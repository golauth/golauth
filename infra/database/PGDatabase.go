package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"github.com/subosito/gotenv"
	"os"
)

var instance *PGDatabase

type PGDatabase struct {
	db *sql.DB
}

func NewPGDatabase() Database {
	var err error
	database := &PGDatabase{}
	connStr := database.createStringConn()
	database.db, err = database.getConnection(connStr)
	if err != nil {
		logrus.Fatal(err)
	}
	err = database.migrate()
	if err != nil {
		logrus.Fatal(err)
	}
	return database
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

func (d PGDatabase) createStringConn() string {
	_ = gotenv.Load()
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUsername, dbPassword, dbName)
}

func (d PGDatabase) getConnection(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("database: could not open connection: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database: could not establish connection: %w", err)
	}
	return db, nil
}

func (d PGDatabase) migrate() error {
	sourceUrl := os.Getenv("MIGRATION_SOURCE_URL")
	logrus.Info("starting migration execution")
	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("database: could not create migration connection: %w", err)
	}
	logrus.Infof("Executing migrations on path: %s", sourceUrl)
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+sourceUrl,
		"postgres", driver,
	)

	if m != nil {
		err = m.Up()
		if err != nil && err.Error() != "no change" {
			return fmt.Errorf("database: error when executing database migration: %w", err)
		}
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
