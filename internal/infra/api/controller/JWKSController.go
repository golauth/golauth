package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/application/keys"
)

// JWKSController publishes the public half of the signing keys as a JWKS
// document so any service can verify a golauth token without calling golauth.
type JWKSController interface {
	JWKS(ctx fiber.Ctx) error
}

type jwksController struct {
	// document is immutable for the life of the process: the key set is fixed
	// at startup, so it is rendered once here.
	document keys.JWKS
}

func NewJWKSController(keySet *keys.KeySet) JWKSController {
	return jwksController{document: keySet.JWKS()}
}

func (c jwksController) JWKS(ctx fiber.Ctx) error {
	ctx.Set(fiber.HeaderCacheControl, "public, max-age=300")
	return ctx.Status(http.StatusOK).JSON(c.document)
}
