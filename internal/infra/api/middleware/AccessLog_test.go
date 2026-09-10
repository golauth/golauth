package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/infra/api/apictx"
	"github.com/stretchr/testify/require"
)

// captureLogs swaps the slog default for one writing JSON into a buffer and
// restores it when the test ends.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return buf
}

func accessLogApp() *fiber.App {
	app := fiber.New()
	app.Use(RequestID())
	app.Use(AccessLog())
	app.Get("/x", func(c fiber.Ctx) error { return c.SendString("ok") })
	return app
}

// linesFor returns the parsed access-log records ("msg":"request") in buf.
func linesFor(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, raw := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if raw == "" {
			continue
		}
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(raw), &m))
		if m["msg"] == "request" {
			out = append(out, m)
		}
	}
	return out
}

// The correlation id on the response header is the same one written to the log,
// so a user reporting a request hands over the exact key to its server-side line.
func TestAccessLogCarriesRequestIDIntoResponseAndLine(t *testing.T) {
	buf := captureLogs(t)

	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(apictx.RequestIDHeader, "trace-xyz-1")
	resp, err := accessLogApp().Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)

	require.Equal(t, "trace-xyz-1", resp.Header.Get(apictx.RequestIDHeader))

	lines := linesFor(t, buf)
	require.Len(t, lines, 1, "exactly one line per request")
	line := lines[0]
	require.Equal(t, "trace-xyz-1", line["request_id"])
	require.Equal(t, "GET", line["method"])
	require.Equal(t, "/x", line["path"])
	require.EqualValues(t, http.StatusOK, line["status"])
	require.Equal(t, "request", line["event"])
	require.Contains(t, line, "duration")
}

// The access log must never contain the Authorization header, its scheme or the
// token value -- it records traffic, not credentials.
func TestAccessLogNeverEmitsAuthorization(t *testing.T) {
	buf := captureLogs(t)

	const token = "super.secret.jwt.value" //nolint:gosec // not a real credential
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	_, err := accessLogApp().Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)

	out := buf.String()
	require.NotEmpty(t, out, "the request was logged")
	require.NotContains(t, out, token)
	require.NotContains(t, out, "Bearer")
	require.NotContains(t, strings.ToLower(out), "authorization")
}

// An unmatched route is a 404; the line records that status, not the default 200.
func TestAccessLogRecordsErrorStatus(t *testing.T) {
	buf := captureLogs(t)

	req, _ := http.NewRequest(http.MethodGet, "/nope", nil)
	_, err := accessLogApp().Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)

	lines := linesFor(t, buf)
	require.Len(t, lines, 1)
	require.EqualValues(t, http.StatusNotFound, lines[0]["status"])
}
