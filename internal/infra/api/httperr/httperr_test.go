package httperr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/application/user"
	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/infra/api/apictx"
	"github.com/stretchr/testify/require"
)

// appReturning builds a one-route app whose handler returns err, wired with the
// Handler under test and a stub that publishes a fixed request id.
func appReturning(err error) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: Handler})
	app.Get("/x", func(c fiber.Ctx) error {
		apictx.SetRequestID(c, "req-123")
		return err
	})
	return app
}

func do(t *testing.T, app *fiber.App) (int, Body, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	raw, _ := io.ReadAll(resp.Body)
	var b Body
	_ = json.Unmarshal(raw, &b)
	return resp.StatusCode, b, string(raw)
}

func TestHandlerMapsDomainSentinels(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"not found", fmt.Errorf("user 7 missing: %w", apperr.ErrNotFound), http.StatusNotFound, "not_found"},
		{"already exists", fmt.Errorf("dup: %w", apperr.ErrAlreadyExists), http.StatusConflict, "already_exists"},
		{"invalid input", fmt.Errorf("bad uuid %q: %w", "abc", apperr.ErrInvalidInput), http.StatusBadRequest, "invalid_input"},
		{"unauthorized", fmt.Errorf("nope: %w", apperr.ErrUnauthorized), http.StatusUnauthorized, "unauthorized"},
		{"forbidden", fmt.Errorf("nope: %w", apperr.ErrForbidden), http.StatusForbidden, "forbidden"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, body, raw := do(t, appReturning(c.err))
			require.Equal(t, c.status, status)
			require.Equal(t, c.code, body.Error.Code)
			require.Equal(t, "req-123", body.Error.RequestID)
			require.NotEmpty(t, body.Error.Message)
			// The wrapped context (ids, "bad uuid", "dup") must not leak.
			require.NotContains(t, raw, "missing")
			require.NotContains(t, raw, "bad uuid")
		})
	}
}

// A repository-style driver error is a generic 500 with no SQL fragment.
func TestHandlerHidesDriverErrorInGeneric500(t *testing.T) {
	driverErr := fmt.Errorf("could not find user by id [7]: %w",
		fmt.Errorf("pq: SELECT * FROM golauth_user failed: connection reset"))

	status, body, raw := do(t, appReturning(driverErr))

	require.Equal(t, http.StatusInternalServerError, status)
	require.Equal(t, "internal_error", body.Error.Code)
	require.Equal(t, "internal server error", body.Error.Message)
	require.Equal(t, "req-123", body.Error.RequestID)
	require.NotContains(t, raw, "pq:")
	require.NotContains(t, raw, "SELECT")
	require.NotContains(t, raw, "golauth_user")
	require.NotContains(t, raw, "connection reset")
}

// A *fiber.Error keeps its status and fixed message but is rendered in the
// contract envelope.
func TestHandlerRendersFiberError(t *testing.T) {
	status, body, _ := do(t, appReturning(fiber.NewError(http.StatusForbidden, "insufficient authority")))
	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, "forbidden", body.Error.Code)
	require.Equal(t, "insufficient authority", body.Error.Message)
}

func TestHandlerRendersValidationFields(t *testing.T) {
	ve := &user.ValidationError{Fields: []user.FieldError{
		{Field: "password", Message: "must be at least 12 characters"},
		{Field: "email", Message: "is not a valid e-mail address"},
	}}
	status, body, _ := do(t, appReturning(ve))

	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "invalid_input", body.Error.Code)
	require.Len(t, body.Fields, 2)
	require.Equal(t, "password", body.Fields[0].Field)
}

func TestHandlerUnmappedErrorIs500(t *testing.T) {
	status, body, raw := do(t, appReturning(fmt.Errorf("something odd happened internally")))
	require.Equal(t, http.StatusInternalServerError, status)
	require.Equal(t, "internal_error", body.Error.Code)
	require.NotContains(t, raw, "something odd")
}
