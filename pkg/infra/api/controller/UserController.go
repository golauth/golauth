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
	var userRole model.UserRoleRequest
	if err := ctx.BodyParser(&userRole); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	err := u.addUserRole.Execute(ctx.UserContext(), userRole.UserID, userRole.RoleID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return ctx.SendStatus(http.StatusCreated)
}
