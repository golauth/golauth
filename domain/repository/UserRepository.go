//go:generate mockgen -source UserRepository.go -destination mock/UserRepository_mock.go -package mock
package repository

import (
	"context"
	"github.com/golauth/golauth/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	Create(ctx context.Context, user entity.User) (entity.User, error)
}
