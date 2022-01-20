package main

import (
	"fmt"
	"github.com/golauth/golauth/api"
	"github.com/golauth/golauth/infra/datasource"
	"log"
	"net/http"
	"os"

	"github.com/subosito/gotenv"
)

func getPortEnv() string {
	_ = gotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return port
}

func main() {
	port := getPortEnv()
	addr := fmt.Sprint(":", port)
	ds, err := datasource.NewDatasource()
	if err != nil {
		log.Fatalf("error when creating database connection: %s", err.Error())
	}
	r := api.NewRouter(ds.GetDB())
	fmt.Println("Server listening on port: ", port)
	log.Fatal(http.ListenAndServe(addr, r.Config()))
}
