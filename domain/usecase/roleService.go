//go:generate mockgen -source roleService.go -destination mock/roleService_mock.go -package mock
package usecase

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
)

type RoleService interface {
	ChangeStatus(ctx context.Context, id uuid.UUID, enabled bool) error
	FindByName(ctx context.Context, name string) (model.RoleResponse, error)
}

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(r repository.RoleRepository) RoleService {
	return roleService{repo: r}
}

func (s roleService) ChangeStatus(ctx context.Context, id uuid.UUID, enabled bool) error {
	exists, err := s.repo.ExistsById(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role with id %s does not exists", id)
	}
	return s.repo.ChangeStatus(ctx, id, enabled)
}

func (s roleService) FindByName(ctx context.Context, name string) (model.RoleResponse, error) {
	role, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return model.RoleResponse{}, err
	}
	return model.RoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		Enabled:      role.Enabled,
		CreationDate: role.CreationDate,
	}, nil
}
