//go:generate mockgen -source RefreshTokenRepository.go -destination mock/RefreshTokenRepository_mock.go -package mock
package repository

import (
	"context"
	"time"

	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	// Create inserts a new refresh-token row and fills the generated ID on the
	// passed entity.
	Create(ctx context.Context, token *entity.RefreshToken) (*entity.RefreshToken, error)
	// FindByHash returns the row whose token_hash matches, or nil (no error)
	// when there is none.
	FindByHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	// Replace marks oldID as rotated: it stamps revoked_at and records newID in
	// replaced_by. It is a no-op on a row that is already revoked.
	Replace(ctx context.Context, oldID, newID uuid.UUID) error
	// Revoke stamps revoked_at on a single row, unless it is already revoked.
	Revoke(ctx context.Context, id uuid.UUID) error
	// RevokeAllForUser revokes every not-yet-revoked token of a user and returns
	// how many rows changed.
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) (int64, error)
	// DeleteExpired removes rows whose expires_at is before the given instant and
	// returns how many were deleted.
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
