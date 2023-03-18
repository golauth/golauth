package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/src/application/user"
	"github.com/golauth/golauth/src/infra/api/controller/model"
	"net/http"
)

type SignupController interface {
	CreateUser(ctx *fiber.Ctx) error
}

type signupController struct {
	createUser user.CreateUser
}

func NewSignupController(createUser user.CreateUser) SignupController {
	return &signupController{createUser: createUser}
}

func (s *signupController) CreateUser(ctx *fiber.Ctx) error {
	var decodedUser model.CreateUserRequest
	if err := ctx.BodyParser(&decodedUser); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	output, err := s.createUser.Execute(ctx.UserContext(), decodedUser.ToEntity())
	if err != nil {
		return err
	}

	return ctx.Status(http.StatusCreated).JSON(output)
}
