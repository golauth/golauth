//go:generate mockgen -source RoleRepository.go -destination mock/RoleRepository_mock.go -package mock
package repository

import (
	"context"
	"github.com/golauth/golauth/domain/entity"
	"github.com/google/uuid"
)

type RoleRepository interface {
	FindByName(ctx context.Context, name string) (*entity.Role, error)
	Create(ctx context.Context, role *entity.Role) (*entity.Role, error)
	Edit(ctx context.Context, role entity.Role) error
	ChangeStatus(ctx context.Context, id uuid.UUID, enabled bool) error
	ExistsById(ctx context.Context, id uuid.UUID) (bool, error)
}
