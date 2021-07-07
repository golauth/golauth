//go:generate mockgen -source datasource.go -destination mock/datasource_mock.go -package mock
package datasource

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/subosito/gotenv"
)

const stringConnBase = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"
const stringConnSchema = " search_path=%s"
const dbSchema = "golauth"

type Datasource interface {
	GetDB() *sql.DB
}

type datasource struct {
	db         *sql.DB
	dbHost     string
	dbPort     string
	dbName     string
	dbUsername string
	dbPassword string
}

func NewDatasource() (Datasource, error) {
	ds := datasource{}
	ds.loadEnvVariables()
	err := ds.createDBConnection()
	if err != nil {
		return nil, err
	}
	return ds, nil
}

func (d datasource) GetDB() *sql.DB {
	return d.db
}

func (d *datasource) createDBConnection() error {
	err := d.validateAndCreateSchema()
	if err != nil {
		return fmt.Errorf("could not create and validate schema: %w", err)
	}
	stringConn := fmt.Sprintf(stringConnBase+stringConnSchema,
		d.dbHost, d.dbPort, d.dbUsername, d.dbPassword, d.dbName, dbSchema)

	d.db, err = sql.Open("postgres", stringConn)
	if err != nil {
		return fmt.Errorf("could not open connection: %w", err)
	}

	err = d.db.Ping()
	if err != nil {
		return fmt.Errorf("could not establish connection: %w", err)
	}

	sourceUrl := os.Getenv("MIGRATION_SOURCE_URL")
	err = d.migration(sourceUrl)
	if err != nil {
		return fmt.Errorf("migration error: %w", err)
	}
	return nil
}

func (d *datasource) loadEnvVariables() {
	_ = gotenv.Load()
	d.dbHost = os.Getenv("DB_HOST")
	d.dbPort = os.Getenv("DB_PORT")
	d.dbName = os.Getenv("DB_NAME")
	d.dbUsername = os.Getenv("DB_USERNAME")
	d.dbPassword = os.Getenv("DB_PASSWORD")
}

func (d datasource) validateAndCreateSchema() error {
	stringConn := fmt.Sprintf(stringConnBase,
		d.dbHost, d.dbPort, d.dbUsername, d.dbPassword, d.dbName)

	dbWithoutSchema, err := sql.Open("postgres", stringConn)
	if err != nil {
		return fmt.Errorf("could not open connection for validate and migrate database: %w", err)
	}
	dbWithoutSchema.QueryRow(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", dbSchema))
	err = dbWithoutSchema.Close()
	if err != nil {
		return fmt.Errorf("could not create database schema: %w", err)
	}
	return nil
}

func (d datasource) migration(sourceUrl string) error {
	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
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
