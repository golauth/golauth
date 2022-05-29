//go:generate mockgen -source AddRole.go -destination mock/AddRole_mock.go -package mock
package role

import (
	"context"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/domain/factory"
	"github.com/golauth/golauth/src/domain/repository"
)

type AddRole interface {
	Execute(ctx context.Context, input *entity.Role) (*entity.Role, error)
}

type addRole struct {
	repo repository.RoleRepository
}

func NewAddRole(repoFactory factory.RepositoryFactory) *addRole {
	return &addRole{repo: repoFactory.NewRoleRepository()}
}

func (uc addRole) Execute(ctx context.Context, input *entity.Role) (*entity.Role, error) {
	role, err := uc.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	return role, nil
}
