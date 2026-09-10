//go:generate mockgen -source Logout.go -destination mock/Logout_mock.go -package mock
package token

import (
	"context"
	"fmt"

	"github.com/golauth/golauth/pkg/application/audit"
	"github.com/golauth/golauth/pkg/domain/factory"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/google/uuid"
)

type Logout interface {
	// Session revokes the single presented refresh token. It is idempotent: an
	// unknown or already-revoked token is not an error.
	Session(ctx context.Context, refreshToken string) error
	// AllSessions revokes every refresh token of the subject.
	AllSessions(ctx context.Context, userID uuid.UUID) error
}

func NewLogout(repoFactory factory.RepositoryFactory) Logout {
	return logout{refreshTokenRepository: repoFactory.NewRefreshTokenRepository()}
}

type logout struct {
	refreshTokenRepository repository.RefreshTokenRepository
}

func (uc logout) Session(ctx context.Context, refreshToken string) error {
	stored, err := uc.refreshTokenRepository.FindByHash(ctx, hashOpaqueToken(refreshToken))
	if err != nil {
		return fmt.Errorf("could not look up refresh token: %w", err)
	}
	if stored == nil {
		return nil
	}
	if err := uc.refreshTokenRepository.Revoke(ctx, stored.ID); err != nil {
		return err
	}
	audit.Event(ctx, audit.Logout, "user_id", stored.UserID.String())
	return nil
}

func (uc logout) AllSessions(ctx context.Context, userID uuid.UUID) error {
	if _, err := uc.refreshTokenRepository.RevokeAllForUser(ctx, userID); err != nil {
		return err
	}
	audit.Event(ctx, audit.LogoutAll, "user_id", userID.String())
	return nil
}
