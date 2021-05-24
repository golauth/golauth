package main

import (
	"fmt"
	"golauth/config/routes"
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

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	pathPrefix = os.Getenv("PATH_PREFIX")
	if pathPrefix == "" {
		pathPrefix = "/golauth"
	}
}

func main() {
	addr := fmt.Sprint(":", port)
	router := mux.NewRouter().PathPrefix(pathPrefix).Subrouter()

	r := routes.NewRoutes(pathPrefix)
	r.RegisterRouter(router)

	fmt.Println("Server listening on port: ", port)
	log.Fatal(http.ListenAndServe(addr, router))
}
