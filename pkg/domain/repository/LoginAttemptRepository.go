//go:generate mockgen -source LoginAttemptRepository.go -destination mock/LoginAttemptRepository_mock.go -package mock
package repository

import (
	"context"
	"time"

	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/google/uuid"
)

type LoginAttemptRepository interface {
	// Get returns the current attempt row for a user, or nil when the user has
	// no recorded failures. A missing row is not an error.
	Get(ctx context.Context, userID uuid.UUID) (*entity.LoginAttempt, error)
	// RegisterFailure increments the failure count, stamps the failure time and
	// stores lockedUntil (a zero value clears the lock). It upserts, so the
	// first failure for a user creates the row.
	RegisterFailure(ctx context.Context, userID uuid.UUID, lockedUntil time.Time) error
	// Reset clears all recorded failures for a user after a successful login.
	Reset(ctx context.Context, userID uuid.UUID) error
}
