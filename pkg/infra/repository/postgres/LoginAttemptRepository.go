package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/google/uuid"
)

type LoginAttemptRepositoryPostgres struct {
	db database.Database
}

func NewLoginAttemptRepository(db database.Database) repository.LoginAttemptRepository {
	return &LoginAttemptRepositoryPostgres{db: db}
}

func (r LoginAttemptRepositoryPostgres) Get(ctx context.Context, userID uuid.UUID) (*entity.LoginAttempt, error) {
	var a entity.LoginAttempt
	var lastFailure, lockedUntil sql.NullTime
	row := r.db.One(ctx,
		`SELECT user_id, failed_count, last_failure, locked_until
		 FROM golauth_login_attempt WHERE user_id = $1`, userID)
	err := row.Scan(&a.UserID, &a.FailedCount, &lastFailure, &lockedUntil)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read login attempt for user %s: %w", userID, err)
	}
	a.LastFailure = lastFailure.Time
	a.LockedUntil = lockedUntil.Time
	return &a, nil
}

func (r LoginAttemptRepositoryPostgres) RegisterFailure(ctx context.Context, userID uuid.UUID, lockedUntil time.Time) error {
	// The column is a bare timestamp (no zone). Store UTC so the value read
	// back -- which lib/pq labels UTC -- denotes the same instant.
	var locked sql.NullTime
	if !lockedUntil.IsZero() {
		locked = sql.NullTime{Time: lockedUntil.UTC(), Valid: true}
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO golauth_login_attempt (user_id, failed_count, last_failure, locked_until)
		 VALUES ($1, 1, now(), $2)
		 ON CONFLICT (user_id) DO UPDATE SET
		     failed_count = golauth_login_attempt.failed_count + 1,
		     last_failure = now(),
		     locked_until = $2`, userID, locked)
	if err != nil {
		return fmt.Errorf("could not register login failure for user %s: %w", userID, err)
	}
	return nil
}

func (r LoginAttemptRepositoryPostgres) Reset(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM golauth_login_attempt WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("could not reset login attempts for user %s: %w", userID, err)
	}
	return nil
}
