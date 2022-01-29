//go:generate mockgen -source UserAuthorityRepository.go -destination mock/UserAuthorityRepository_mock.go -package mock
package repository

import (
	"context"
	"github.com/google/uuid"
)

type UserAuthorityRepository interface {
	FindAuthoritiesByUserID(ctx context.Context, userId uuid.UUID) ([]string, error)
}
