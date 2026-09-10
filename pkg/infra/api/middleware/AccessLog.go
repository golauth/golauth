package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/infra/api/apictx"
	"github.com/golauth/golauth/pkg/infra/api/httperr"
)

// AccessLog emits exactly one structured line per request: method, path,
// status, duration, correlation id, client ip, and the authenticated subject
// when the request carried a valid token.
//
// It deliberately never reads the Authorization header, a body or a token: an
// access log is a traffic record, not a place credentials should ever land.
//
// Register it right after RequestID and before recover and the security
// middleware, so the correlation id is already assigned and even a rejected or
// panicking request is still logged. It reads the final status from the error
// the chain returns (via httperr.Status) because that status is only written to
// the response later, by the central error handler.
func AccessLog() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now()
		err := ctx.Next()

		status := ctx.Response().StatusCode()
		if err != nil {
			status = httperr.Status(err)
		}

		attrs := []any{
			slog.String("event", "request"),
			slog.String("request_id", apictx.RequestID(ctx)),
			slog.String("method", ctx.Method()),
			slog.String("path", ctx.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("client_ip", ctx.IP()),
		}
		if c, ok := apictx.ClaimsFromContext(ctx); ok {
			attrs = append(attrs, slog.String("subject", c.Subject))
		}

		slog.LogAttrs(ctx.Context(), levelForStatus(status), "request", toAttrs(attrs)...)
		return err
	}
}

// levelForStatus keeps a routine 2xx/3xx at info, surfaces client errors at
// warn and server errors at error, so an alert can trigger on level alone.
func levelForStatus(status int) slog.Level {
	switch {
	case status >= fiber.StatusInternalServerError:
		return slog.LevelError
	case status >= fiber.StatusBadRequest:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func toAttrs(args []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args))
	for _, a := range args {
		if attr, ok := a.(slog.Attr); ok {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}
