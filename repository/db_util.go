package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io/ioutil"
	"log"
	"os"
)

var (
	pgC               testcontainers.Container
	mockScanError     = errors.New("scan error")
	mockDBClosedError = errors.New("sql: database is closed")
	mockUpdateError   = errors.New("exec update error")
)

const (
	testDbHost          = "localhost"
	testDbName          = "golauth_test"
	testDbUser          = "test"
	testDbPassword      = "test"
	testPostgresSvcPort = "5432"
)

func Up(configRootPath bool) context.Context {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:12-alpine",
		ExposedPorts: []string{testPostgresSvcPort},
		Env: map[string]string{
			"POSTGRES_DB":       testDbName,
			"POSTGRES_USER":     testDbUser,
			"POSTGRES_PASSWORD": testDbPassword,
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(testPostgresSvcPort),
			wait.ForSQL(testPostgresSvcPort, "postgres", func(port nat.Port) string {
				return fmt.Sprintf(
					"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
					testDbHost,
					port.Port(),
					testDbUser,
					testDbPassword,
					testDbName,
				)
			}),
		),
	}
	var err error
	pgC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}
	testDbPort, _ := pgC.MappedPort(ctx, testPostgresSvcPort)
	log.Printf("Database started as port: %s", testDbPort)

	setEnv(testDbPort, configRootPath)
	return ctx
}

func setEnv(testDbPort nat.Port, configRootPath bool) {
	_ = os.Setenv("DB_HOST", testDbHost)
	_ = os.Setenv("DB_PORT", testDbPort.Port())
	_ = os.Setenv("DB_NAME", testDbName)
	_ = os.Setenv("DB_USERNAME", testDbUser)
	_ = os.Setenv("DB_PASSWORD", testDbPassword)
	_ = os.Setenv("MIGRATION_SOURCE_URL", getMigrationsPath(configRootPath))
}

func getMigrationsPath(configRootPath bool) string {
	return fmt.Sprintf("%s/migrations", getRootPath(configRootPath))
}

func Down(ctx context.Context) {
	err := pgC.Terminate(ctx)
	if err != nil {
		log.Println(err.Error())
	}
}

func Cleanup(db *sql.DB, configRootPath bool) {
	c, ioErr := ioutil.ReadFile(getCleanupScript(configRootPath))
	if ioErr != nil {
		log.Fatal(ioErr)
	}
	sqlFile := string(c)
	_, err := db.Exec(sqlFile)
	if err != nil {
		log.Fatal(err)
	}
}

func newDBMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func getCleanupScript(configRootPath bool) string {
	return fmt.Sprintf("%s/util/test/cleanup.sql", getRootPath(configRootPath))
}

func getRootPath(configRootPath bool) string {
	if configRootPath {
		return os.Getenv("ROOT_PATH")
	}
	return "."
}
