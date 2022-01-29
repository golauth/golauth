//go:generate mockgen -source FindRoleByName.go -destination mock/FindRoleByName_mock.go -package mock
package role

import (
	"context"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/repository"
)

type FindRoleByName interface {
	Execute(ctx context.Context, name string) (*entity.Role, error)
}

func NewFindRoleByName(repo repository.RoleRepository) FindRoleByName {
	return findRoleByName{repo: repo}
}

type findRoleByName struct {
	repo repository.RoleRepository
}

func (uc findRoleByName) Execute(ctx context.Context, name string) (*entity.Role, error) {
	role, err := uc.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return role, nil
}
