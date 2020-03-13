package db

import (
	"database/sql"
	"fmt"
	"golauth/util"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/subosito/gotenv"
)

var (
	db         *sql.DB
	err        error
	dbhost     string
	dbport     string
	dbname     string
	dbusername string
	dbpassword string
)

const stringConnBase = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"
const stringConnSchema = " search_path=%s"
const dbschema = "golauth"

func init() {
	_ = gotenv.Load()
	dbhost = os.Getenv("DB_HOST")
	dbport = os.Getenv("DB_PORT")
	dbname = os.Getenv("DB_NAME")
	dbusername = os.Getenv("DB_USERNAME")
	dbpassword = os.Getenv("DB_PASSWORD")

	validateAndCreateSchema()
	stringConn := fmt.Sprintf(stringConnBase+stringConnSchema,
		dbhost, dbport, dbusername, dbpassword, dbname, dbschema)

	db, err = sql.Open("postgres", stringConn)
	util.LogFatal(err)

	err = db.Ping()
	util.LogFatal(err)

	sourceUrl := os.Getenv("MIGRATION_SOURCE_URL")
	migration(sourceUrl)
}

func validateAndCreateSchema() {
	stringConn := fmt.Sprintf(stringConnBase,
		dbhost, dbport, dbusername, dbpassword, dbname)

	dbWithoutSchema, err := sql.Open("postgres", stringConn)
	util.LogFatal(err)
	dbWithoutSchema.QueryRow(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", dbschema))
	dbWithoutSchema.Close()
}

func GetDatasource() *sql.DB {
	return db
}

func migration(sourceUrl string) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	util.LogFatal(err)
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
