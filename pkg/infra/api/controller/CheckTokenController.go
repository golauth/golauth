package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/token"
)

type CheckTokenController interface {
	CheckToken(ctx fiber.Ctx) error
}

type checkTokenController struct {
	validateToken token.ValidateToken
}

func NewCheckTokenController(validateToken token.ValidateToken) CheckTokenController {
	return checkTokenController{validateToken: validateToken}
}

func (c checkTokenController) CheckToken(ctx fiber.Ctx) error {
	t, err := token.ExtractToken(ctx.Get("Authorization"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	_, err = c.validateToken.Execute(t)
	if err != nil {
		return fiber.NewError(http.StatusUnauthorized, err.Error())
	}
	return ctx.SendStatus(http.StatusNoContent)
}
