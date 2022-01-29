package main

import (
	"fmt"
	"github.com/golauth/golauth/infra/api"
	"github.com/golauth/golauth/infra/database"
	"github.com/golauth/golauth/infra/factory"
	"log"
	"net/http"
	"os"

	"github.com/subosito/gotenv"
)

func getPortEnv() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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
	r := api.NewRouter(rf)
	fmt.Println("Server listening on port: ", port)
	log.Fatal(http.ListenAndServe(addr, r.Config()))
}
