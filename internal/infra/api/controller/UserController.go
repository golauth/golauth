package controller

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/application/user"
	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/infra/api/controller/model"
	"github.com/google/uuid"
)

type UserController struct {
	findById    user.FindUserById
	addUserRole user.AddUserRole
}

func NewUserController(findById user.FindUserById, addUserRole user.AddUserRole) UserController {
	return UserController{findById: findById, addUserRole: addUserRole}
}

func (u UserController) FindById(ctx fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("invalid user id %q: %w", ctx.Params("id"), apperr.ErrInvalidInput)
	}
	data, err := u.findById.Execute(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(http.StatusOK).JSON(model.NewUserResponseFromEntity(data))
}

func (u UserController) AddRole(ctx fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("invalid user id %q: %w", ctx.Params("id"), apperr.ErrInvalidInput)
	}
	var userRole model.UserRoleRequest
	if err := ctx.Bind().Body(&userRole); err != nil {
		return fmt.Errorf("invalid request body: %w", apperr.ErrInvalidInput)
	}
	// The path is the source of truth: it is what the authorization layer saw.
	// A body naming a different user is a mismatch, not an override.
	if userRole.UserID != uuid.Nil && userRole.UserID != userID {
		return fmt.Errorf("body user id does not match the request path: %w", apperr.ErrInvalidInput)
	}
	if err := u.addUserRole.Execute(ctx.Context(), userID, userRole.RoleID); err != nil {
		return err
	}

	return ctx.SendStatus(http.StatusCreated)
}
