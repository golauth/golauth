package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/application/token"
	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/domain/repository"
	"github.com/golauth/golauth/internal/infra/api/apictx"
	"github.com/golauth/golauth/internal/infra/api/controller/model"
	"github.com/google/uuid"
)

// ErrContentTypeNotSupported is the fixed message for a login attempt with an
// unusable content type. It is a 405.
var ErrContentTypeNotSupported = errors.New("content-type not supported")

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
		return fmt.Errorf("invalid request body: %w", apperr.ErrInvalidInput)
	}

	if userLogin == (model.UserLoginRequest{}) {
		return fmt.Errorf("missing credentials: %w", apperr.ErrInvalidInput)
	}

	clientIP, userAgent := ctx.IP(), ctx.Get("User-Agent")
	output, err := s.generateToken.Execute(ctx.Context(), userLogin.Username, userLogin.Password, clientIP, userAgent)
	if err != nil {
		// Deliberately vague: an unknown user and a wrong password are
		// indistinguishable in the response.
		return fmt.Errorf("invalid credentials: %w", apperr.ErrUnauthorized)
	}

	return ctx.Status(http.StatusOK).JSON(model.NewTokenResponseFromEntity(output))
}

// Refresh is public: it takes an opaque refresh token, rotates it and returns a
// new pair. Every failure is a flat 401.
func (s tokenController) Refresh(ctx fiber.Ctx) error {
	req, err := bindRefreshToken(ctx)
	if err != nil {
		return err
	}
	output, err := s.refreshAccessToken.Execute(ctx.Context(), req.RefreshToken, ctx.IP(), ctx.Get("User-Agent"))
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", apperr.ErrUnauthorized)
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
		return err
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// LogoutAll is authenticated: it revokes every refresh token of the token
// subject.
func (s tokenController) LogoutAll(ctx fiber.Ctx) error {
	claims, ok := apictx.ClaimsFromContext(ctx)
	if !ok {
		return fmt.Errorf("no authenticated principal: %w", apperr.ErrUnauthorized)
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return fmt.Errorf("token subject is not a uuid: %w", apperr.ErrUnauthorized)
	}
	if err := s.logout.AllSessions(ctx.Context(), userID); err != nil {
		return err
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func bindRefreshToken(ctx fiber.Ctx) (model.RefreshTokenRequest, error) {
	var req model.RefreshTokenRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return req, fmt.Errorf("invalid request body: %w", apperr.ErrInvalidInput)
	}
	if req.RefreshToken == "" {
		return req, fmt.Errorf("missing refresh_token: %w", apperr.ErrInvalidInput)
	}
	return req, nil
}
