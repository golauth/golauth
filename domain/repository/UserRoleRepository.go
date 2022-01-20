//go:generate mockgen -source UserRoleRepository.go -destination mock/UserRoleRepository_mock.go -package mock
package repository

import "github.com/google/uuid"

type UserRoleRepository interface {
	AddUserRole(userId uuid.UUID, roleId uuid.UUID) error
}
