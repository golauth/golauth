//go:generate mockgen -source EditRole.go -destination mock/EditRole_mock.go -package mock
package role

import (
	"context"
	"fmt"

	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/domain/entity"
	"github.com/golauth/golauth/internal/domain/repository"
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
		return fmt.Errorf("role %s: %w", id, apperr.ErrNotFound)
	}
	if id != input.ID {
		return fmt.Errorf("path id %s and body id %s do not match: %w", id, input.ID, apperr.ErrInvalidInput)
	}
	return uc.repo.Edit(ctx, input)
}
