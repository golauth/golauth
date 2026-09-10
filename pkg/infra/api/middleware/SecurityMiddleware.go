package middleware

import (
	"strings"

	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/infra/api/apictx"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
)

// ClaimsFromContext returns the claims published by SecurityMiddleware for the
// current request. The second result is false when the request did not go
// through authentication, which authorization middlewares must treat as a
// denial rather than as an anonymous-but-allowed request.
//
// It forwards to apictx.ClaimsFromContext, which owns the Locals key so that
// controllers can read the claims without importing this package.
func ClaimsFromContext(ctx fiber.Ctx) (*model.Claims, bool) {
	return apictx.ClaimsFromContext(ctx)
}

type SecurityMiddleware struct {
	validateToken token.ValidateToken
	publicURI     map[string]bool
}

func NewSecurityMiddleware(validateToken token.ValidateToken, pathPrefix string) *SecurityMiddleware {
	return &SecurityMiddleware{
		validateToken: validateToken,
		publicURI: map[string]bool{
			pathPrefix + "/token":                 true,
			pathPrefix + "/token/refresh":         true,
			pathPrefix + "/check_token":           true,
			pathPrefix + "/signup":                true,
			pathPrefix + "/.well-known/jwks.json": true,
		},
	}
}

// Apply authenticates every request whose path is not explicitly public.
//
// This middleware MUST be registered before any route it is meant to protect:
// the fiber router dispatches the first matching stack entry and stops, so a
// route registered earlier is served by its handler without ever reaching this
// one. That ordering mistake was GHSA-p34g-m47x-q2m4.
func (s *SecurityMiddleware) Apply() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if !s.isPrivateURI(ctx.Path()) {
			return ctx.Next()
		}

		bearerTk := ctx.Get(fiber.HeaderAuthorization, "")
		t, err := token.ExtractToken(bearerTk)
		if err != nil {
			// Fixed, uninformative message on purpose: a detailed reason here
			// helps an attacker map the system.
			return fiber.NewError(http.StatusUnauthorized, "unauthorized")
		}
		claims, err := s.validateToken.Execute(t)
		if err != nil {
			return fiber.NewError(http.StatusUnauthorized, "unauthorized")
		}

		apictx.SetClaims(ctx, claims)
		return ctx.Next()
	}
}

// isPrivateURI decides on the request path alone. It used to compare the full
// absolute URI ("http://host:8080/auth/token") against path-shaped keys, so no
// entry ever matched.
func (s *SecurityMiddleware) isPrivateURI(path string) bool {
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	_, contains := s.publicURI[path]
	return !contains
}
