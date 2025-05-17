package main

import (
	"fmt"
	"github.com/golauth/golauth/src/infra/api"
	"github.com/golauth/golauth/src/infra/database"
	"github.com/golauth/golauth/src/infra/factory"
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
	port := getPortEnv()
	addr := fmt.Sprint(":", port)
	db := database.NewPGDatabase()
	defer db.Close()
	rf := factory.NewPostgresRepositoryFactory(db)
	app := api.NewRouter(rf)
	fmt.Println("Server listening on port: ", port)
	log.Fatal(app.Config().Listen(addr))
}
