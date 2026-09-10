package controller

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

// readinessTimeout bounds the database ping the readiness probe makes, so a
// wedged database turns into a fast 503 rather than a hanging probe.
const readinessTimeout = 2 * time.Second

// HealthPinger is the readiness probe's view of the database: just a bounded
// liveness check, nothing on the request path.
type HealthPinger interface {
	Ping(ctx context.Context) error
}

// HealthController serves the two probes. They answer different questions and
// must fail independently: liveness is "is the process up" and never touches a
// dependency; readiness is "can it serve traffic right now" and checks the
// database and the signing key.
type HealthController struct {
	live  fiber.Handler
	ready fiber.Handler
}

// NewHealthController wires the probes. keyLoaded reports whether the JWT
// signing key is present; without it the service can authenticate requests but
// cannot mint tokens, so it is not ready.
func NewHealthController(db HealthPinger, keyLoaded func() bool) HealthController {
	return HealthController{
		live: healthcheck.New(healthcheck.Config{ResponseFormat: healthcheck.FormatJSON}),
		ready: healthcheck.New(healthcheck.Config{
			ResponseFormat: healthcheck.FormatJSON,
			Probe: func(c fiber.Ctx) bool {
				ctx, cancel := context.WithTimeout(c.Context(), readinessTimeout)
				defer cancel()
				if err := db.Ping(ctx); err != nil {
					slog.WarnContext(ctx, "readiness: database ping failed", "err", err.Error())
					return false
				}
				if !keyLoaded() {
					slog.WarnContext(ctx, "readiness: signing key not loaded")
					return false
				}
				return true
			},
		}),
	}
}

// Live answers 200 as long as the process can serve a request. An orchestrator
// restarts the pod when this fails, so it must not depend on the database.
func (h HealthController) Live(c fiber.Ctx) error { return h.live(c) }

// Ready answers 200 when the database is reachable and the signing key is
// loaded, 503 otherwise. An orchestrator pulls the pod out of the load balancer
// when this fails, without restarting it.
func (h HealthController) Ready(c fiber.Ctx) error { return h.ready(c) }
