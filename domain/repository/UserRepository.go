//go:generate mockgen -source UserRepository.go -destination mock/UserRepository_mock.go -package mock
package repository

import (
	"github.com/google/uuid"
	"golauth/domain/entity"
)

type UserRepository interface {
	FindByUsername(username string) (entity.User, error)
	FindByID(id uuid.UUID) (entity.User, error)
	Create(user entity.User) (entity.User, error)
}
