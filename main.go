package main

import (
	"fmt"
	"golauth/config/datasource"
	"golauth/config/routes"
	"golauth/util"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/subosito/gotenv"
)

var (
	port       string
	pathPrefix string
)

func init() {
	_ = gotenv.Load()
	util.LoadKeyEnv()

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	pathPrefix = os.Getenv("PATH_PREFIX")
	if pathPrefix == "" {
		pathPrefix = "/auth"
	}
}

func main() {
	addr := fmt.Sprint(":", port)
	router := mux.NewRouter().PathPrefix(pathPrefix).Subrouter()

	db, err := datasource.CreateDBConnection()
	if err != nil {
		log.Fatalf("error when creating database connection: %s", err.Error())
	}

	r := routes.NewRoutes(pathPrefix, db)
	r.RegisterRouter(router)

	fmt.Println("Server listening on port: ", port)
	log.Fatal(http.ListenAndServe(addr, router))
}
