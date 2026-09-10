// Package apictx holds the request-scoped values the API stashes in fiber
// Locals. It exists so both the middleware that publishes the authenticated
// claims and the controllers that read them can share one unexported key
// without importing each other.
package apictx

import (
	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/token/claims"
)

// claimsKey is the Locals key for the current request's validated claims. The
// unexported type makes it impossible for another package to overwrite by
// accident.
type claimsKey struct{}

// SetClaims publishes the validated claims of the current request. Only the
// authentication middleware should call it.
func SetClaims(ctx fiber.Ctx, c *claims.Claims) {
	ctx.Locals(claimsKey{}, c)
}

// ClaimsFromContext returns the claims published by the authentication
// middleware. The second result is false when the request never went through
// authentication, which authorization code must treat as a denial rather than
// as an anonymous-but-allowed request.
func ClaimsFromContext(ctx fiber.Ctx) (*claims.Claims, bool) {
	c, ok := ctx.Locals(claimsKey{}).(*claims.Claims)
	return c, ok && c != nil
}
