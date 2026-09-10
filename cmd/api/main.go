package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

const (
	defaultPort = "8080"
	// defaultShutdownTimeout bounds the drain: in-flight requests get this long
	// to finish after SIGTERM before connections are cut. Override with
	// SERVER_SHUTDOWN_TIMEOUT.
	defaultShutdownTimeout = 15 * time.Second
	// healthProbeTimeout bounds the self-probe the -healthcheck flag makes.
	healthProbeTimeout = 3 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("golauth exited", "err", err.Error())
		os.Exit(1)
	}
}

// run wires the process and blocks until the server stops -- cleanly on
// SIGINT/SIGTERM, or with an error if the listener fails. It is kept small and
// error-returning so the shutdown path in serve() is what carries the logic.
func run() error {
	_ = gotenv.Load()
	logging.Setup()

	healthcheckFlag := flag.Bool("healthcheck", false, "probe the local /health/live endpoint and exit (for Docker HEALTHCHECK)")
	flag.Parse()

	port := getPortEnv()
	if *healthcheckFlag {
		return selfHealthCheck(port)
	}

	keySet, err := keys.Load()
	if err != nil {
		return fmt.Errorf("loading jwt signing key: %w", err)
	}

	db := database.NewPGDatabase()
	rf := factory.NewPostgresRepositoryFactory(db)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Purge expired refresh-token rows now and on a timer so the table stays
	// bounded. REFRESH_TOKEN_CLEANUP_INTERVAL (a Go duration) overrides the 1h
	// default; a deployment preferring an external cron can set it very long.
	// It stops when ctx is cancelled by the shutdown signal.
	token.StartRefreshTokenCleanup(ctx, rf.NewRefreshTokenRepository(), cleanupInterval())

	app := api.NewRouter(rf, keySet, db).Config()
	return serve(ctx, app, ":"+port, shutdownTimeout(), db)
}

// shutdownTimeout reads SERVER_SHUTDOWN_TIMEOUT (a Go duration), falling back to
// the default on anything unset or invalid.
func shutdownTimeout() time.Duration {
	if v := os.Getenv("SERVER_SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		slog.Warn("invalid SERVER_SHUTDOWN_TIMEOUT, using default",
			"value", v, "default", defaultShutdownTimeout.String())
	}
	return defaultShutdownTimeout
}

// serve runs app until ctx is cancelled or the listener fails, then drains
// in-flight connections within timeout and closes db. Factored out of run() so
// a test can drive it with an ephemeral port and a mock database.
func serve(ctx context.Context, app *fiber.App, addr string, timeout time.Duration, db interface{ Close() }) error {
	listenErr := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", addr)
		listenErr <- app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true})
	}()

	select {
	case err := <-listenErr:
		db.Close()
		if err != nil {
			return fmt.Errorf("listener: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	slog.Info("shutdown signal received; draining", "timeout", timeout.String())
	shutdownErr := app.ShutdownWithTimeout(timeout)
	db.Close()
	if shutdownErr != nil {
		return fmt.Errorf("graceful shutdown: %w", shutdownErr)
	}
	slog.Info("server drained and stopped")
	return nil
}

// selfHealthCheck is the body of `golauth -healthcheck`: a plain GET against
// the local liveness endpoint. The runtime image is scratch-like with no curl
// or wget, so the binary probes itself.
func selfHealthCheck(port string) error {
	client := &http.Client{Timeout: healthProbeTimeout}
	resp, err := client.Get("http://127.0.0.1:" + port + "/health/live")
	if err != nil {
		return fmt.Errorf("health probe: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health probe: status %d", resp.StatusCode)
	}
	return nil
}

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
