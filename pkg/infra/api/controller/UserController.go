package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/pkg/application/user"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/google/uuid"
	"net/http"
)

type UserController struct {
	findById    user.FindUserById
	addUserRole user.AddUserRole
}

func NewUserController(findById user.FindUserById, addUserRole user.AddUserRole) UserController {
	return UserController{findById: findById, addUserRole: addUserRole}
}

func (u UserController) FindById(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	data, err := u.findById.Execute(ctx.UserContext(), id)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return ctx.Status(http.StatusOK).JSON(model.NewUserResponseFromEntity(data))
}

func (u UserController) AddRole(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	var userRole model.UserRoleRequest
	if err := ctx.BodyParser(&userRole); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	// The path is the source of truth: it is what the authorization layer saw.
	// A body naming a different user is a mismatch, not an override.
	if userRole.UserID != uuid.Nil && userRole.UserID != userID {
		return fiber.NewError(http.StatusBadRequest, "userId does not match the request path")
	}
	err = u.addUserRole.Execute(ctx.UserContext(), userID, userRole.RoleID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return ctx.SendStatus(http.StatusCreated)
}
