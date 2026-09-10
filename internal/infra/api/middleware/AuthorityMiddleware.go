package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/application/token/claims"
)

// AdminAuthority is the authority granted by the ADMIN role, as seeded by the
// initial_data migration.
const AdminAuthority = "ADMIN"

// RequireAuthority rejects requests whose token does not carry the given
// authority. Authenticating is not enough to reach an administrative endpoint:
// any self-registered account holds a valid token, so without this check the
// role-management routes remain a privilege-escalation sink.
func RequireAuthority(authority string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		c, ok := ClaimsFromContext(ctx)
		if !ok || !hasAuthority(c, authority) {
			return fiber.NewError(http.StatusForbidden, "insufficient authority")
		}
		return ctx.Next()
	}
}

// RequireSelfOrAuthority allows a caller to act on its own user, identified by
// the token subject, and otherwise demands the given authority. The user id is
// read from the named route parameter.
func RequireSelfOrAuthority(param string, authority string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		c, ok := ClaimsFromContext(ctx)
		if !ok {
			return fiber.NewError(http.StatusForbidden, "insufficient authority")
		}
		if c.Subject != "" && c.Subject == ctx.Params(param) {
			return ctx.Next()
		}
		if hasAuthority(c, authority) {
			return ctx.Next()
		}
		return fiber.NewError(http.StatusForbidden, "insufficient authority")
	}
}

func hasAuthority(c *claims.Claims, authority string) bool {
	for _, a := range c.Authorities {
		if a == authority {
			return true
		}
	}
	return false
}
