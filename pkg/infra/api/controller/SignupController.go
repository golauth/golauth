package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/user"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
)

type SignupController interface {
	CreateUser(ctx fiber.Ctx) error
}

type signupController struct {
	createUser user.CreateUser
}

func NewSignupController(createUser user.CreateUser) SignupController {
	return &signupController{createUser: createUser}
}

func (s *signupController) CreateUser(ctx fiber.Ctx) error {
	var decodedUser model.CreateUserRequest
	if err := ctx.Bind().Body(&decodedUser); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	output, err := s.createUser.Execute(ctx.Context(), decodedUser.ToEntity())
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Serialize through UserResponse rather than the entity: the entity
	// carries the bcrypt hash, and signup answers unauthenticated callers.
	return ctx.Status(http.StatusCreated).JSON(model.NewUserResponseFromEntity(output))
}
