package datasource

import (
	"database/sql"
	"fmt"
	"log"
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

func CreateDBConnection() *sql.DB {
	var err error
	_ = gotenv.Load()
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_NAME")
	dbUsername = os.Getenv("DB_USERNAME")
	dbPassword = os.Getenv("DB_PASSWORD")

	validateAndCreateSchema()
	stringConn := fmt.Sprintf(stringConnBase+stringConnSchema,
		dbHost, dbPort, dbUsername, dbPassword, dbName, dbschema)

	db, err = sql.Open("postgres", stringConn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	sourceUrl := os.Getenv("MIGRATION_SOURCE_URL")
	migration(sourceUrl)
	return db
}

func validateAndCreateSchema() {
	stringConn := fmt.Sprintf(stringConnBase,
		dbHost, dbPort, dbUsername, dbPassword, dbName)

	dbWithoutSchema, err := sql.Open("postgres", stringConn)
	if err != nil {
		log.Fatal(err)
	}
	dbWithoutSchema.QueryRow(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", dbschema))
	err = dbWithoutSchema.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func migration(sourceUrl string) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+sourceUrl,
		"postgres", driver,
	)

	if m != nil {
		err = m.Up()
		if err != nil && err.Error() != "no change" {
			log.Fatal(err)
		}
	}
}
