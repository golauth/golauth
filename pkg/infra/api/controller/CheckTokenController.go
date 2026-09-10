package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/infra/api/apictx"
)

type CheckTokenController interface {
	CheckToken(ctx fiber.Ctx) error
	Me(ctx fiber.Ctx) error
}

type checkTokenController struct {
	validateToken token.ValidateToken
}

func NewCheckTokenController(validateToken token.ValidateToken) CheckTokenController {
	return checkTokenController{validateToken: validateToken}
}

// CheckToken is RFC 7662-ish introspection: it verifies the bearer token and,
// on success, returns 200 with the claims it just verified so a consumer does
// not have to decode the JWT itself. Offline verification via the JWKS endpoint
// is the preferred integration; this is the convenience.
func (c checkTokenController) CheckToken(ctx fiber.Ctx) error {
	t, err := token.ExtractToken(ctx.Get("Authorization"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	claims, err := c.validateToken.Execute(t)
	if err != nil {
		return fiber.NewError(http.StatusUnauthorized, err.Error())
	}
	return ctx.Status(http.StatusOK).JSON(claims)
}

// Me is the ergonomic variant for a browser or gateway: it returns the claims
// the security middleware already verified and published for this request. It
// touches neither the token parser nor the database.
func (c checkTokenController) Me(ctx fiber.Ctx) error {
	claims, ok := apictx.ClaimsFromContext(ctx)
	if !ok {
		return fiber.NewError(http.StatusUnauthorized)
	}
	return ctx.Status(http.StatusOK).JSON(claims)
}
