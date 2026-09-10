package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/keys"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/infra/api"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/golauth/golauth/pkg/infra/factory"
	"github.com/golauth/golauth/pkg/infra/logging"

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
		slog.Warn("invalid REFRESH_TOKEN_CLEANUP_INTERVAL, using default",
			"value", v, "default", token.DefaultCleanupInterval.String())
	}
	return token.DefaultCleanupInterval
}

func main() {
	_ = gotenv.Load()
	// Configure the process-wide logger before anything else logs.
	logging.Setup()

	port := getPortEnv()
	addr := fmt.Sprint(":", port)

	keySet, err := keys.Load()
	if err != nil {
		slog.Error("loading jwt signing key", "err", err.Error())
		os.Exit(1)
	}

	db := database.NewPGDatabase()
	defer db.Close()
	rf := factory.NewPostgresRepositoryFactory(db)

	// Purge expired refresh-token rows now and on a timer so the table stays
	// bounded. REFRESH_TOKEN_CLEANUP_INTERVAL (a Go duration) overrides the 1h
	// default; a deployment preferring an external cron can set it very long.
	token.StartRefreshTokenCleanup(context.Background(), rf.NewRefreshTokenRepository(), cleanupInterval())

	app := api.NewRouter(rf, keySet)
	slog.Info("server listening", "port", port)
	if err := app.Config().Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
		slog.Error("server stopped", "err", err.Error())
		os.Exit(1)
	}
}
