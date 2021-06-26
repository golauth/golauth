package datasource

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/subosito/gotenv"
)

var (
	db         *sql.DB
	dbHost     string
	dbPort     string
	dbName     string
	dbUsername string
	dbPassword string
)

const stringConnBase = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"
const stringConnSchema = " search_path=%s"
const dbschema = "golauth"

func CreateDBConnection() (*sql.DB, error) {
	var err error
	_ = gotenv.Load()
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_NAME")
	dbUsername = os.Getenv("DB_USERNAME")
	dbPassword = os.Getenv("DB_PASSWORD")

	err = validateAndCreateSchema()
	if err != nil {
		return nil, err
	}
	stringConn := fmt.Sprintf(stringConnBase+stringConnSchema,
		dbHost, dbPort, dbUsername, dbPassword, dbName, dbschema)

	db, err = sql.Open("postgres", stringConn)
	if err != nil {
		return nil, fmt.Errorf("could not open connection: %w", err)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not stabilish connection: %w", err)
	}

	sourceUrl := os.Getenv("MIGRATION_SOURCE_URL")
	err = migration(sourceUrl)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func validateAndCreateSchema() error {
	stringConn := fmt.Sprintf(stringConnBase,
		dbHost, dbPort, dbUsername, dbPassword, dbName)

	dbWithoutSchema, err := sql.Open("postgres", stringConn)
	if err != nil {
		return fmt.Errorf("could not open connection for validate and migrate database: %w", err)
	}
	dbWithoutSchema.QueryRow(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", dbschema))
	err = dbWithoutSchema.Close()
	if err != nil {
		return fmt.Errorf("could not create database schema: %w", err)
	}
	return nil
}

func migration(sourceUrl string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration connection: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+sourceUrl,
		"postgres", driver,
	)

	if m != nil {
		err = m.Up()
		if err != nil && err.Error() != "no change" {
			return fmt.Errorf("error when executing database migration: %w", err)
		}
	}
	return nil
}
