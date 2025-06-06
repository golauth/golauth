package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/pkg/application/token"
	"net/http"
)

type SecurityMiddleware struct {
	validateToken token.ValidateToken
	publicURI     map[string]bool
}

func NewSecurityMiddleware(validateToken token.ValidateToken, pathPrefix string) *SecurityMiddleware {
	return &SecurityMiddleware{
		validateToken: validateToken,
		publicURI: map[string]bool{
			pathPrefix + "/token":       true,
			pathPrefix + "/check_token": true,
			pathPrefix + "/signup":      true,
		},
	}
}

func (s *SecurityMiddleware) Apply() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		requestURI := ctx.Request().URI().String()
		if s.isPrivateURI(requestURI) {
			fmt.Println(ctx.GetReqHeaders())
			bearerTk := ctx.Get("Authorization", "")
			t, err := token.ExtractToken(bearerTk)
			if err != nil {
				return fiber.NewError(http.StatusInternalServerError, err.Error())
			}
			err = s.validateToken.Execute(t)
			if err != nil {
				return fiber.NewError(http.StatusUnauthorized, err.Error())
			}
		}
		return ctx.Next()
	}
}

func (s *SecurityMiddleware) isPrivateURI(requestURI string) bool {
	_, contains := s.publicURI[requestURI]
	return !contains
}
