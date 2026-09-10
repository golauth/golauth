package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
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
		claims, ok := ClaimsFromContext(ctx)
		if !ok || !hasAuthority(claims, authority) {
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
		claims, ok := ClaimsFromContext(ctx)
		if !ok {
			return fiber.NewError(http.StatusForbidden, "insufficient authority")
		}
		if claims.Subject != "" && claims.Subject == ctx.Params(param) {
			return ctx.Next()
		}
		if hasAuthority(claims, authority) {
			return ctx.Next()
		}
		return fiber.NewError(http.StatusForbidden, "insufficient authority")
	}
}

func hasAuthority(claims *model.Claims, authority string) bool {
	for _, a := range claims.Authorities {
		if a == authority {
			return true
		}
	}
	return false
}
