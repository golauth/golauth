package main

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/src/infra/api"
	"github.com/golauth/golauth/src/infra/database"
	"github.com/golauth/golauth/src/infra/factory"
	"github.com/golauth/golauth/src/infra/monitoring"
	"log"
	"os"

	"github.com/subosito/gotenv"
)

const defaultPort = "8080"

func getPortEnv() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return port
}

func main() {
	_ = gotenv.Load()

	shutdown, err := monitoring.InitOTELProvider()
	if err != nil {
		log.Fatal(err)
	}

	port := getPortEnv()
	addr := fmt.Sprint(":", port)
	db := database.NewPGDatabase()
	defer func() {
		db.Close()
		if err := shutdown(context.Background()); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	rf := factory.NewPostgresRepositoryFactory(db)
	app := api.NewRouter(rf)
	fmt.Println("Server listening on port: ", port)
	log.Fatal(app.Config().Listen(addr))
}
