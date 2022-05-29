//go:generate mockgen -source AddUserRole.go -destination mock/AddUserRole_mock.go -package mock
package user

import (
	"context"
	"github.com/golauth/golauth/src/domain/repository"
	"github.com/google/uuid"
)

type AddUserRole interface {
	Execute(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
}

func NewAddUserRole(repo repository.UserRoleRepository) AddUserRole {
	return addUserRole{repo: repo}
}

type addUserRole struct {
	repo repository.UserRoleRepository
}

func (uc addUserRole) Execute(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	return uc.repo.AddUserRole(ctx, userID, roleID)
}
