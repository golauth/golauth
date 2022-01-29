//go:generate mockgen -source ChangeRoleStatus.go -destination mock/ChangeRoleStatus_mock.go -package mock
package role

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/domain/repository"
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
		return fmt.Errorf("role with id %s does not exists", id)
	}
	return uc.repo.ChangeStatus(ctx, id, enabled)
}
