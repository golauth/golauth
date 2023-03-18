package controller

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/src/application/role"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/domain/factory"
	"github.com/golauth/golauth/src/infra/api/controller/model"
	"github.com/google/uuid"
	"net/http"
)

type RoleController struct {
	addRole          role.AddRole
	editRole         role.EditRole
	changeRoleStatus role.ChangeRoleStatus
	findByName       role.FindRoleByName
}

func NewRoleController(repoFactory factory.RepositoryFactory) RoleController {
	return RoleController{
		addRole:          role.NewAddRole(repoFactory),
		editRole:         role.NewEditRole(repoFactory.NewRoleRepository()),
		changeRoleStatus: role.NewChangeRoleStatus(repoFactory.NewRoleRepository()),
		findByName:       role.NewFindRoleByName(repoFactory.NewRoleRepository()),
	}
}

func (c RoleController) Create(ctx *fiber.Ctx) error {
	var data model.RoleRequest
	fmt.Println(ctx.GetReqHeaders())
	if err := ctx.BodyParser(&data); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	input := entity.NewRole(data.Name, data.Description)
	output, err := c.addRole.Execute(ctx.UserContext(), input)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return ctx.Status(http.StatusCreated).JSON(output)
}

func (c RoleController) Edit(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("cannot cast %s to uuid: %w", id, err)
	}
	var data model.RoleRequest
	if err := ctx.BodyParser(&data); err != nil {
		return err
	}
	err = c.editRole.Execute(ctx.UserContext(), id, data.ToEntity())
	if err != nil {
		return err
	}
	ctx.Status(http.StatusOK)
	return ctx.JSON(data)
}

func (c RoleController) ChangeStatus(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("cannot cast %s to uuid: %v", id, err))
	}
	var data model.RoleChangeStatus
	if err := ctx.BodyParser(&data); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	err = c.changeRoleStatus.Execute(ctx.UserContext(), id, data.Enabled)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func (c RoleController) FindByName(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	data, err := c.findByName.Execute(ctx.UserContext(), name)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return ctx.Status(http.StatusOK).JSON(data)
}
