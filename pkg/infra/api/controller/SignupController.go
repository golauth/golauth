package controller

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/user"
	"github.com/golauth/golauth/pkg/domain/repository"
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
		return signupError(ctx, err)
	}

	// Serialize through UserResponse rather than the entity: the entity
	// carries the bcrypt hash, and signup answers unauthenticated callers.
	return ctx.Status(http.StatusCreated).JSON(model.NewUserResponseFromEntity(output))
}

// signupError turns a use-case error into an honest status code: 400 with the
// field list for invalid input, 409 for a duplicate username or e-mail, and 500
// for anything genuinely unexpected.
func signupError(ctx fiber.Ctx, err error) error {
	var ve *user.ValidationError
	if errors.As(err, &ve) {
		return ctx.Status(http.StatusBadRequest).JSON(ve)
	}
	if errors.Is(err, repository.ErrUserAlreadyExists) {
		return fiber.NewError(http.StatusConflict, "username or e-mail already registered")
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}
