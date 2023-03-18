package controller

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/src/application/token"
	"github.com/golauth/golauth/src/domain/repository"
	"github.com/golauth/golauth/src/infra/api/controller/model"
	"net/http"
)

var (
	ErrContentTypeNotSupported = errors.New("content-type not supported")
	ErrMissingBodyData         = errors.New("missing body data")
)

type TokenController interface {
	Token(ctx *fiber.Ctx) error
}

type tokenController struct {
	userRepository          repository.UserRepository
	userAuthorityRepository repository.UserAuthorityRepository
	generateToken           token.GenerateToken
}

func NewTokenController(
	userRepository repository.UserRepository,
	userAuthorityRepository repository.UserAuthorityRepository,
	generateToken token.GenerateToken) TokenController {
	return tokenController{
		userRepository:          userRepository,
		userAuthorityRepository: userAuthorityRepository,
		generateToken:           generateToken,
	}
}

func (s tokenController) Token(ctx *fiber.Ctx) error {
	var userLogin model.UserLoginRequest

	contentType := ctx.Get("Content-Type")
	if contentType != "application/json" && contentType != "application/x-www-form-urlencoded" {
		return fiber.NewError(http.StatusMethodNotAllowed, ErrContentTypeNotSupported.Error())
	}

	if err := ctx.BodyParser(&userLogin); err != nil {
		return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("json decoder error: %v", err))
	}

	if userLogin == (model.UserLoginRequest{}) {
		return fiber.NewError(http.StatusBadRequest, ErrMissingBodyData.Error())
	}

	output, err := s.generateToken.Execute(ctx.UserContext(), userLogin.Username, userLogin.Password)
	if err != nil {
		return fiber.NewError(http.StatusUnauthorized)
	}

	return ctx.Status(http.StatusOK).JSON(&model.TokenResponse{AccessToken: output.AccessToken})
}
