// Package audit emits the security-relevant transitions of the service --
// logins, refreshes, logouts, user and role changes -- as structured records
// with a stable "event" key so an operator can alert on them.
//
// Audit events are deliberately separate from the per-request access log: the
// access log records that traffic happened, audit events record that a security
// decision was made. They share one logger (log/slog) so there is still only
// one output stream to ship.
//
// It imports nothing from pkg/infra: the layering guard forbids it, and the
// standard library slog default logger is all it needs.
package audit

import (
	"context"
	"log/slog"
)

// Stable event names. Alert rules and dashboards key on these exact strings, so
// treat a rename as a breaking change.
const (
	LoginSucceeded    = "login_succeeded"
	LoginFailed       = "login_failed"
	TokenRefreshed    = "token_refreshed"
	RefreshTokenReuse = "refresh_token_reuse"
	RefreshDenied     = "refresh_denied"
	Logout            = "logout"
	LogoutAll         = "logout_all"
	UserCreated       = "user_created"
	RoleCreated       = "role_created"
	RoleGranted       = "role_granted"
	RoleStatusChanged = "role_status_changed"
)

// Logger is the sink for audit records. It defaults to slog.Default, which
// logging.Setup configures; a test swaps it to capture output.
var Logger = func() *slog.Logger { return slog.Default() }

// Event writes one audit record at info level. name must be one of the
// constants above; args are alternating key/value pairs, exactly as
// slog.Logger.Info takes them.
func Event(ctx context.Context, name string, args ...any) {
	Logger().With(slog.String("event", name)).InfoContext(ctx, "audit event", args...)
}

// Warn writes an audit record at warn level, for a transition that is itself a
// red flag -- a rotated refresh token presented again, say.
func Warn(ctx context.Context, name string, args ...any) {
	Logger().With(slog.String("event", name)).WarnContext(ctx, "audit event", args...)
}
