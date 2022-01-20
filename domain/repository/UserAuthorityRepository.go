//go:generate mockgen -source UserAuthorityRepository.go -destination mock/UserAuthorityRepository_mock.go -package mock
package repository

import "github.com/google/uuid"

type UserAuthorityRepository interface {
	FindAuthoritiesByUserID(userId uuid.UUID) ([]string, error)
}
