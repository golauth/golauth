package middleware

import (
	"net/http"
	"testing"

	"github.com/cristalhq/jwt/v3"
	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/infra/api/apictx"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withClaims stands in for SecurityMiddleware, publishing claims the way it
// does. Passing nil models a request that never went through authentication.
func withClaims(claims *model.Claims) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if claims != nil {
			apictx.SetClaims(ctx, claims)
		}
		return ctx.Next()
	}
}

func ok(ctx fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) }

func TestRequireAuthority(t *testing.T) {
	tests := map[string]struct {
		claims   *model.Claims
		expected int
	}{
		"admin is allowed":            {&model.Claims{Authorities: []string{"USER", "ADMIN"}}, http.StatusOK},
		"plain user is denied":        {&model.Claims{Authorities: []string{"USER"}}, http.StatusForbidden},
		"no authorities is denied":    {&model.Claims{}, http.StatusForbidden},
		"unauthenticated is denied":   {nil, http.StatusForbidden},
		"similar authority is denied": {&model.Claims{Authorities: []string{"ADMINISTRATOR"}}, http.StatusForbidden},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/roles", withClaims(tc.claims), RequireAuthority(AdminAuthority), ok)

			req, _ := http.NewRequest("GET", "/roles", nil)
			resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tc.expected, resp.StatusCode)
		})
	}
}

func TestRequireSelfOrAuthority(t *testing.T) {
	const self = "37fe41b4-24bf-4da9-9124-615cc72865a5"
	const other = "9c3f2b21-0000-4da9-9124-615cc72865a5"

	tests := map[string]struct {
		claims   *model.Claims
		target   string
		expected int
	}{
		"owner reaches itself": {
			&model.Claims{Authorities: []string{"USER"}, StandardClaims: standardWithSubject(self)}, self, http.StatusOK,
		},
		"non admin cannot reach another user": {
			&model.Claims{Authorities: []string{"USER"}, StandardClaims: standardWithSubject(self)}, other, http.StatusForbidden,
		},
		"admin reaches any user": {
			&model.Claims{Authorities: []string{"ADMIN"}, StandardClaims: standardWithSubject(self)}, other, http.StatusOK,
		},
		// An empty subject must never be treated as matching an empty or
		// absent path parameter.
		"empty subject is denied": {
			&model.Claims{Authorities: []string{"USER"}}, other, http.StatusForbidden,
		},
		"unauthenticated is denied": {nil, self, http.StatusForbidden},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/users/:id", withClaims(tc.claims), RequireSelfOrAuthority("id", AdminAuthority), ok)

			req, _ := http.NewRequest("GET", "/users/"+tc.target, nil)
			resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tc.expected, resp.StatusCode)
		})
	}
}

func standardWithSubject(subject string) jwt.StandardClaims {
	return jwt.StandardClaims{Subject: subject}
}
