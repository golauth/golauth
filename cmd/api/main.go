package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/keys"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/infra/api"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/golauth/golauth/pkg/infra/factory"
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

func cleanupInterval() time.Duration {
	if v := os.Getenv("REFRESH_TOKEN_CLEANUP_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		logrus.Warnf("invalid REFRESH_TOKEN_CLEANUP_INTERVAL=%q, using default %s", v, token.DefaultCleanupInterval)
	}
	return token.DefaultCleanupInterval
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

	// Purge expired refresh-token rows now and on a timer so the table stays
	// bounded. REFRESH_TOKEN_CLEANUP_INTERVAL (a Go duration) overrides the 1h
	// default; a deployment preferring an external cron can set it very long.
	token.StartRefreshTokenCleanup(context.Background(), rf.NewRefreshTokenRepository(), cleanupInterval())

	app := api.NewRouter(rf, keySet)
	fmt.Println("Server listening on port: ", port)
	log.Fatal(app.Config().Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}))
}
