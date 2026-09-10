package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/api/apictx"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/google/uuid"
)

var (
	ErrContentTypeNotSupported = errors.New("content-type not supported")
	ErrMissingBodyData         = errors.New("missing body data")
	ErrMissingRefreshToken     = errors.New("missing refresh_token")
)

type TokenController interface {
	Token(ctx fiber.Ctx) error
	Refresh(ctx fiber.Ctx) error
	Logout(ctx fiber.Ctx) error
	LogoutAll(ctx fiber.Ctx) error
}

type tokenController struct {
	userRepository          repository.UserRepository
	userAuthorityRepository repository.UserAuthorityRepository
	generateToken           token.GenerateToken
	refreshAccessToken      token.RefreshAccessToken
	logout                  token.Logout
}

func NewTokenController(
	userRepository repository.UserRepository,
	userAuthorityRepository repository.UserAuthorityRepository,
	generateToken token.GenerateToken,
	refreshAccessToken token.RefreshAccessToken,
	logout token.Logout) TokenController {
	return tokenController{
		userRepository:          userRepository,
		userAuthorityRepository: userAuthorityRepository,
		generateToken:           generateToken,
		refreshAccessToken:      refreshAccessToken,
		logout:                  logout,
	}
}

func (s tokenController) Token(ctx fiber.Ctx) error {
	var userLogin model.UserLoginRequest

	contentType := ctx.Get("Content-Type")
	if contentType != "application/json" && contentType != "application/x-www-form-urlencoded" {
		return fiber.NewError(http.StatusMethodNotAllowed, ErrContentTypeNotSupported.Error())
	}

	if err := ctx.Bind().Body(&userLogin); err != nil {
		return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("json decoder error: %v", err))
	}

	if userLogin == (model.UserLoginRequest{}) {
		return fiber.NewError(http.StatusBadRequest, ErrMissingBodyData.Error())
	}

	clientIP, userAgent := ctx.IP(), ctx.Get("User-Agent")
	output, err := s.generateToken.Execute(ctx.Context(), userLogin.Username, userLogin.Password, clientIP, userAgent)
	if err != nil {
		return fiber.NewError(http.StatusUnauthorized)
	}

	return ctx.Status(http.StatusOK).JSON(model.NewTokenResponseFromEntity(output))
}

// Refresh is public: it takes an opaque refresh token, rotates it and returns a
// new pair. Every failure is a flat 401 with no body.
func (s tokenController) Refresh(ctx fiber.Ctx) error {
	req, err := bindRefreshToken(ctx)
	if err != nil {
		return err
	}
	output, err := s.refreshAccessToken.Execute(ctx.Context(), req.RefreshToken, ctx.IP(), ctx.Get("User-Agent"))
	if err != nil {
		return fiber.NewError(http.StatusUnauthorized)
	}
	return ctx.Status(http.StatusOK).JSON(model.NewTokenResponseFromEntity(output))
}

// Logout is authenticated: it revokes the presented refresh token. The current
// access token keeps working until it expires -- the documented trade-off of
// not maintaining an access-token denylist.
func (s tokenController) Logout(ctx fiber.Ctx) error {
	req, err := bindRefreshToken(ctx)
	if err != nil {
		return err
	}
	if err := s.logout.Session(ctx.Context(), req.RefreshToken); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "could not revoke refresh token")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// LogoutAll is authenticated: it revokes every refresh token of the token
// subject.
func (s tokenController) LogoutAll(ctx fiber.Ctx) error {
	claims, ok := apictx.ClaimsFromContext(ctx)
	if !ok {
		return fiber.NewError(http.StatusUnauthorized)
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return fiber.NewError(http.StatusUnauthorized)
	}
	if err := s.logout.AllSessions(ctx.Context(), userID); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "could not revoke refresh tokens")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func bindRefreshToken(ctx fiber.Ctx) (model.RefreshTokenRequest, error) {
	var req model.RefreshTokenRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return req, fiber.NewError(http.StatusBadRequest, fmt.Sprintf("json decoder error: %v", err))
	}
	if req.RefreshToken == "" {
		return req, fiber.NewError(http.StatusBadRequest, ErrMissingRefreshToken.Error())
	}
	return req, nil
}
