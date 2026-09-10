package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/infra/api"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/golauth/golauth/pkg/infra/factory"
	"github.com/golauth/golauth/pkg/infra/keys"
	"github.com/sirupsen/logrus"

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
	port := getPortEnv()
	addr := fmt.Sprint(":", port)

	keySet, err := keys.Load()
	if err != nil {
		logrus.Fatalf("loading jwt signing key: %v", err)
	}

	db := database.NewPGDatabase()
	defer db.Close()
	rf := factory.NewPostgresRepositoryFactory(db)
	app := api.NewRouter(rf, keySet)
	fmt.Println("Server listening on port: ", port)
	log.Fatal(app.Config().Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}))
}
