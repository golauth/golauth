//go:generate mockgen -source RoleRepository.go -destination mock/RoleRepository_mock.go -package mock
package repository

import (
	"github.com/google/uuid"
	"golauth/domain/entity"
)

type RoleRepository interface {
	FindByName(name string) (entity.Role, error)
	Create(role entity.Role) (entity.Role, error)
	Edit(role entity.Role) error
	ChangeStatus(id uuid.UUID, enabled bool) error
	ExistsById(id uuid.UUID) (bool, error)
}
