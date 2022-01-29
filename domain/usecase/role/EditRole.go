package role

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
)

type EditRole interface {
	Execute(ctx context.Context, id uuid.UUID, req model.RoleRequest) error
}

type editRole struct {
	repo repository.RoleRepository
}

func NewEditRole(repo repository.RoleRepository) EditRole {
	return editRole{repo: repo}
}

func (uc editRole) Execute(ctx context.Context, id uuid.UUID, req model.RoleRequest) error {
	exists, err := uc.repo.ExistsById(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role with id %s does not exists", id)
	}
	if id != req.ID {
		return fmt.Errorf("path id[%s] and object_id[%s] does not match", id, req.ID)
	}
	data := entity.Role{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}
	return uc.repo.Edit(ctx, data)
}
