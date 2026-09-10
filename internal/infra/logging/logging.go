// Package logging owns the process-wide logger configuration. The whole service
// logs through log/slog; this package is the single place its output format,
// level and destination are decided, so nothing else needs to know how logging
// is wired.
package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// requestIDKey is the context key under which the API stashes the correlation
// id so every log record made with that context -- an access line, an audit
// event, a stray warning from a use case -- can be tied back to one request.
type requestIDKey struct{}

// ContextWithRequestID returns a copy of ctx carrying id. The RequestID
// middleware calls it once per request; everything downstream just passes the
// context along.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the correlation id carried by ctx, or "".
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// WithContext wraps h so that every record it writes is decorated with the
// request id carried by the log call's context (see ContextWithRequestID),
// unless the caller already passed a request_id attribute. Setup applies it;
// a test that needs the same behaviour over its own buffer calls it directly.
func WithContext(h slog.Handler) slog.Handler {
	return contextHandler{h}
}

// contextHandler decorates every record with the context's request id, unless
// the caller already set one explicitly.
type contextHandler struct{ slog.Handler }

func (h contextHandler) Handle(ctx context.Context, rec slog.Record) error {
	id := RequestIDFromContext(ctx)
	if id != "" && !hasAttr(rec, "request_id") {
		rec = rec.Clone()
		rec.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, rec)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}

func hasAttr(rec slog.Record, key string) bool {
	found := false
	rec.Attrs(func(a slog.Attr) bool {
		if a.Key == key {
			found = true
			return false
		}
		return true
	})
	return found
}

// Environment variables read at setup.
const (
	// EnvAppEnv selects the output format: "production" emits JSON, anything
	// else emits human-readable text.
	EnvAppEnv = "APP_ENV"
	// EnvLogLevel is the minimum level to emit: debug, info, warn or error.
	EnvLogLevel = "LOG_LEVEL"
)

// Setup builds the logger from the environment and installs it as the slog
// default, so every slog.Info / slog.Warn / slog.Error call in the process --
// and every audit and access-log record -- goes through it. Call it once, first
// thing in main.
//
//	APP_ENV=production -> JSON lines on stdout (for a log shipper)
//	otherwise          -> readable text on stderr (for a terminal)
//	LOG_LEVEL          -> debug|info|warn|error, default info
func Setup() {
	slog.SetDefault(New())
}

// New returns the logger Setup installs, without installing it. Tests that need
// to assert on output build their own logger instead.
func New() *slog.Logger {
	opts := &slog.HandlerOptions{Level: level(os.Getenv(EnvLogLevel))}
	var h slog.Handler
	if IsProduction() {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stderr, opts)
	}
	return slog.New(WithContext(h))
}

// IsProduction reports whether APP_ENV names the production environment. It is
// the same test the signing-key loader uses to fail closed.
func IsProduction() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(EnvAppEnv)), "production")
}

// level maps a LOG_LEVEL string to a slog.Level, defaulting to info on an empty
// or unrecognised value.
func level(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
