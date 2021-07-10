package main

import (
	"fmt"
	"golauth/config/datasource"
	"golauth/config/routes"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/subosito/gotenv"
)

func getServerEnv() (string, string) {
	_ = gotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	pathPrefix := os.Getenv("PATH_PREFIX")
	if pathPrefix == "" {
		pathPrefix = "/auth"
	}
	return port, pathPrefix
}

func main() {
	port, pathPrefix := getServerEnv()
	addr := fmt.Sprint(":", port)
	router := mux.NewRouter().PathPrefix(pathPrefix).Subrouter()
	ds, err := datasource.NewDatasource()
	if err != nil {
		log.Fatalf("error when creating database connection: %s", err.Error())
	}
	r := routes.NewRouter(pathPrefix, ds.GetDB())
	r.RegisterRoutes(router)
	fmt.Println("Server listening on port: ", port)
	log.Fatal(http.ListenAndServe(addr, router))
}
