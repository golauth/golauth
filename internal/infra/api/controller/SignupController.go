package controller

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/application/user"
	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/infra/api/controller/model"
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
		return fmt.Errorf("invalid request body: %w", apperr.ErrInvalidInput)
	}
	// The use case returns a *user.ValidationError for a bad field and
	// repository.ErrUserAlreadyExists (which wraps apperr.ErrAlreadyExists) for a
	// duplicate; the central handler maps both.
	output, err := s.createUser.Execute(ctx.Context(), decodedUser.ToEntity())
	if err != nil {
		return err
	}

	// Serialize through UserResponse rather than the entity: the entity
	// carries the bcrypt hash, and signup answers unauthenticated callers.
	return ctx.Status(http.StatusCreated).JSON(model.NewUserResponseFromEntity(output))
}
