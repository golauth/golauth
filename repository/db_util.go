package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io/ioutil"
	"log"
	"os"
)

var (
	pgC testcontainers.Container
)

const (
	testDbHost          = "localhost"
	testDbName          = "golauth_test"
	testDbUser          = "test"
	testDbPassword      = "test"
	testPostgresSvcPort = "5432"
)

func Up() context.Context {
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

	_ = os.Setenv("DB_HOST", testDbHost)
	_ = os.Setenv("DB_PORT", testDbPort.Port())
	_ = os.Setenv("DB_NAME", testDbName)
	_ = os.Setenv("DB_USERNAME", testDbUser)
	_ = os.Setenv("DB_PASSWORD", testDbPassword)
	_ = os.Setenv("MIGRATION_SOURCE_URL", getMigrationsPath())
	return ctx
}

func getMigrationsPath() string {
	return fmt.Sprintf("%s/migrations", getRootPath())
}

func Down(ctx context.Context) {
	err := pgC.Terminate(ctx)
	if err != nil {
		log.Println(err.Error())
	}
}

func Cleanup(db *sql.DB) {
	c, ioErr := ioutil.ReadFile(getCleanupScript())
	if ioErr != nil {
		log.Fatal(ioErr)
	}
	sqlFile := string(c)
	_, err := db.Exec(sqlFile)
	if err != nil {
		log.Fatal(err)
	}
}

func getCleanupScript() string {
	return fmt.Sprintf("%s/util/test/cleanup.sql", getRootPath())
}

func getRootPath() string {
	return os.Getenv("ROOT_PATH")
}
