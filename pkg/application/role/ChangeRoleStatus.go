//go:generate mockgen -source ChangeRoleStatus.go -destination mock/ChangeRoleStatus_mock.go -package mock
package role

import (
	"context"
	"fmt"

	"github.com/golauth/golauth/pkg/domain/apperr"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/google/uuid"
)

type ChangeRoleStatus interface {
	Execute(ctx context.Context, id uuid.UUID, enabled bool) error
}

type changeRoleStatus struct {
	repo repository.RoleRepository
}

func NewChangeRoleStatus(repo repository.RoleRepository) ChangeRoleStatus {
	return changeRoleStatus{repo: repo}
}

func (uc changeRoleStatus) Execute(ctx context.Context, id uuid.UUID, enabled bool) error {
	exists, err := uc.repo.ExistsById(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role %s: %w", id, apperr.ErrNotFound)
	}
	return uc.repo.ChangeStatus(ctx, id, enabled)
}
