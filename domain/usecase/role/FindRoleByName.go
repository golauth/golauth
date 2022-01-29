//go:generate mockgen -source FindRoleByName.go -destination mock/FindRoleByName_mock.go -package mock
package role

import (
	"context"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/api/controller/model"
)

type FindRoleByName interface {
	Execute(ctx context.Context, name string) (model.RoleResponse, error)
}

func NewFindRoleByName(repo repository.RoleRepository) FindRoleByName {
	return findRoleByName{repo: repo}
}

type findRoleByName struct {
	repo repository.RoleRepository
}

func (uc findRoleByName) Execute(ctx context.Context, name string) (model.RoleResponse, error) {
	role, err := uc.repo.FindByName(ctx, name)
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
