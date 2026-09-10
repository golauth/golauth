//go:generate mockgen -source UserRoleRepository.go -destination mock/UserRoleRepository_mock.go -package mock
package repository

import (
	"context"
	"github.com/google/uuid"
)

type UserRoleRepository interface {
	AddUserRole(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error
}
