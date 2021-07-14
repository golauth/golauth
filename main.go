package main

import (
	"fmt"
	"golauth/api"
	datasource2 "golauth/infrastructure/datasource"
	"log"
	"net/http"
	"os"

	"github.com/subosito/gotenv"
)

func getServerEnv() string {
	_ = gotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return port
}

func main() {
	port := getServerEnv()
	addr := fmt.Sprint(":", port)
	ds, err := datasource2.NewDatasource()
	if err != nil {
		log.Fatalf("error when creating database connection: %s", err.Error())
	}
	r := api.NewRouter(ds.GetDB())
	fmt.Println("Server listening on port: ", port)
	log.Fatal(http.ListenAndServe(addr, r.Config()))
}
