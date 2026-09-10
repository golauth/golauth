package middleware

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/infra/api/apictx"
	"github.com/stretchr/testify/require"
)

func requestIDApp() *fiber.App {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/x", func(c fiber.Ctx) error {
		return c.SendString(apictx.RequestID(c))
	})
	return app
}

func TestRequestIDGeneratedWhenAbsent(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	resp, err := requestIDApp().Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)

	header := resp.Header.Get(apictx.RequestIDHeader)
	require.NotEmpty(t, header)
	require.Len(t, header, 36) // uuid v4 canonical form
}

func TestRequestIDReusesShortInboundValue(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(apictx.RequestIDHeader, "trace-abc-1")
	resp, err := requestIDApp().Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)

	require.Equal(t, "trace-abc-1", resp.Header.Get(apictx.RequestIDHeader))
}

func TestRequestIDReplacesOversizedInboundValue(t *testing.T) {
	huge := make([]byte, 500)
	for i := range huge {
		huge[i] = 'a'
	}
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(apictx.RequestIDHeader, string(huge))
	resp, err := requestIDApp().Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)

	got := resp.Header.Get(apictx.RequestIDHeader)
	require.NotEqual(t, string(huge), got)
	require.Len(t, got, 36)
}
