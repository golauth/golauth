//go:generate mockgen -source EditRole.go -destination mock/EditRole_mock.go -package mock
package role

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/domain/repository"
	"github.com/google/uuid"
)

type EditRole interface {
	Execute(ctx context.Context, id uuid.UUID, input *entity.Role) error
}

type editRole struct {
	repo repository.RoleRepository
}

func NewEditRole(repo repository.RoleRepository) EditRole {
	return editRole{repo: repo}
}

func (uc editRole) Execute(ctx context.Context, id uuid.UUID, input *entity.Role) error {
	exists, err := uc.repo.ExistsById(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role with id %s does not exists", id)
	}
	if id != input.ID {
		return fmt.Errorf("path id[%s] and object_id[%s] does not match", id, input.ID)
	}
	return uc.repo.Edit(ctx, input)
}
