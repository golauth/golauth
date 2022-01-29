//go:generate mockgen -source AddRole.go -destination mock/AddRole_mock.go -package mock
package role

import (
	"context"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/factory"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/api/controller/model"
)

type AddRole interface {
	Execute(ctx context.Context, input model.RoleRequest) (*model.RoleResponse, error)
}

type addRole struct {
	repo repository.RoleRepository
}

func NewAddRole(repoFactory factory.RepositoryFactory) *addRole {
	return &addRole{repo: repoFactory.NewRoleRepository()}
}

func (uc addRole) Execute(ctx context.Context, input model.RoleRequest) (*model.RoleResponse, error) {
	data := entity.Role{
		Name:        input.Name,
		Description: input.Description,
		Enabled:     true,
	}
	role, err := uc.repo.Create(ctx, data)
	if err != nil {
		return nil, err
	}
	return &model.RoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		Enabled:      role.Enabled,
		CreationDate: role.CreationDate,
	}, nil
}
