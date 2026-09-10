package controller

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/application/role"
	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/domain/entity"
	"github.com/golauth/golauth/internal/domain/factory"
	"github.com/golauth/golauth/internal/infra/api/controller/model"
	"github.com/google/uuid"
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

func (c RoleController) Create(ctx fiber.Ctx) error {
	var data model.RoleRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return fmt.Errorf("invalid request body: %w", apperr.ErrInvalidInput)
	}

	output, err := c.addRole.Execute(ctx.Context(), entity.NewRole(data.Name, data.Description))
	if err != nil {
		return err
	}

	return ctx.Status(http.StatusCreated).JSON(output)
}

func (c RoleController) Edit(ctx fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		// Print the value that actually failed to parse, not the zero uuid.
		return fmt.Errorf("invalid role id %q: %w", ctx.Params("id"), apperr.ErrInvalidInput)
	}
	var data model.RoleRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return fmt.Errorf("invalid request body: %w", apperr.ErrInvalidInput)
	}
	if err := c.editRole.Execute(ctx.Context(), id, data.ToEntity()); err != nil {
		return err
	}
	return ctx.Status(http.StatusOK).JSON(data)
}

func (c RoleController) ChangeStatus(ctx fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("invalid role id %q: %w", ctx.Params("id"), apperr.ErrInvalidInput)
	}
	var data model.RoleChangeStatus
	if err := ctx.Bind().Body(&data); err != nil {
		return fmt.Errorf("invalid request body: %w", apperr.ErrInvalidInput)
	}
	if err := c.changeRoleStatus.Execute(ctx.Context(), id, data.Enabled); err != nil {
		return err
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func (c RoleController) FindByName(ctx fiber.Ctx) error {
	data, err := c.findByName.Execute(ctx.Context(), ctx.Params("name"))
	if err != nil {
		return err
	}

	return ctx.Status(http.StatusOK).JSON(data)
}
