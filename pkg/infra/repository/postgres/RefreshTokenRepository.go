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

type RefreshTokenRepositoryPostgres struct {
	db database.Database
}

func NewRefreshTokenRepository(db database.Database) repository.RefreshTokenRepository {
	return &RefreshTokenRepositoryPostgres{db: db}
}

func (r RefreshTokenRepositoryPostgres) Create(ctx context.Context, token *entity.RefreshToken) (*entity.RefreshToken, error) {
	// Bare timestamp columns: store UTC so the value read back denotes the same
	// instant, matching LoginAttemptRepository.
	err := r.db.One(ctx,
		`INSERT INTO golauth_refresh_token (user_id, token_hash, issued_at, expires_at, user_agent, client_ip)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		token.UserID, token.TokenHash, token.IssuedAt.UTC(), token.ExpiresAt.UTC(),
		nullString(token.UserAgent), nullString(token.ClientIP),
	).Scan(&token.ID)
	if err != nil {
		return nil, fmt.Errorf("could not create refresh token for user %s: %w", token.UserID, err)
	}
	return token, nil
}

func (r RefreshTokenRepositoryPostgres) FindByHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var (
		t          entity.RefreshToken
		revokedAt  sql.NullTime
		replacedBy uuid.NullUUID
		userAgent  sql.NullString
		clientIP   sql.NullString
	)
	row := r.db.One(ctx,
		`SELECT id, user_id, token_hash, issued_at, expires_at, revoked_at, replaced_by, user_agent, client_ip
		 FROM golauth_refresh_token WHERE token_hash = $1`, tokenHash)
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.IssuedAt, &t.ExpiresAt,
		&revokedAt, &replacedBy, &userAgent, &clientIP)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read refresh token: %w", err)
	}
	if revokedAt.Valid {
		t.RevokedAt = &revokedAt.Time
	}
	if replacedBy.Valid {
		id := replacedBy.UUID
		t.ReplacedBy = &id
	}
	t.UserAgent = userAgent.String
	t.ClientIP = clientIP.String
	return &t, nil
}

func (r RefreshTokenRepositoryPostgres) Replace(ctx context.Context, oldID, newID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE golauth_refresh_token SET revoked_at = now(), replaced_by = $2
		 WHERE id = $1 AND revoked_at IS NULL`, oldID, newID)
	if err != nil {
		return fmt.Errorf("could not rotate refresh token %s: %w", oldID, err)
	}
	return nil
}

func (r RefreshTokenRepositoryPostgres) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE golauth_refresh_token SET revoked_at = now()
		 WHERE id = $1 AND revoked_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("could not revoke refresh token %s: %w", id, err)
	}
	return nil
}

func (r RefreshTokenRepositoryPostgres) RevokeAllForUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	res, err := r.db.Exec(ctx,
		`UPDATE golauth_refresh_token SET revoked_at = now()
		 WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	if err != nil {
		return 0, fmt.Errorf("could not revoke refresh tokens for user %s: %w", userID, err)
	}
	return res.RowsAffected()
}

func (r RefreshTokenRepositoryPostgres) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.Exec(ctx,
		`DELETE FROM golauth_refresh_token WHERE expires_at < $1`, before.UTC())
	if err != nil {
		return 0, fmt.Errorf("could not delete expired refresh tokens: %w", err)
	}
	return res.RowsAffected()
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
