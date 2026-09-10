//go:generate mockgen -source RefreshAccessToken.go -destination mock/RefreshAccessToken_mock.go -package mock
package token

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golauth/golauth/pkg/application/audit"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/factory"
	"github.com/golauth/golauth/pkg/domain/repository"
)

// ErrInvalidRefreshToken is the single failure the refresh endpoint reports.
// Unknown, expired, revoked, rotated and disabled-user all collapse to it so the
// endpoint cannot be used to probe token or account state.
var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type RefreshAccessToken interface {
	Execute(ctx context.Context, refreshToken, clientIP, userAgent string) (*entity.Token, error)
}

func NewRefreshAccessToken(repoFactory factory.RepositoryFactory, jwtToken GenerateJwtToken, cfg Config) RefreshAccessToken {
	refreshRepo := repoFactory.NewRefreshTokenRepository()
	return refreshAccessToken{
		refreshTokenRepository: refreshRepo,
		userRepository:         repoFactory.NewUserRepository(),
		issuer: tokenIssuer{
			refreshTokenRepository:  refreshRepo,
			userAuthorityRepository: repoFactory.NewUserAuthorityRepository(),
			jwtToken:                jwtToken,
			cfg:                     cfg,
		},
	}
}

type refreshAccessToken struct {
	refreshTokenRepository repository.RefreshTokenRepository
	userRepository         repository.UserRepository
	issuer                 tokenIssuer
}

func (uc refreshAccessToken) Execute(ctx context.Context, presented, clientIP, userAgent string) (*entity.Token, error) {
	now := time.Now()
	stored, err := uc.refreshTokenRepository.FindByHash(ctx, hashOpaqueToken(presented))
	if err != nil {
		return nil, fmt.Errorf("could not look up refresh token: %w", err)
	}
	if stored == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Reuse detection: a token that was already rotated away is presented again.
	// Someone holds a copy they should not. Burn every session for this user --
	// the standard defence, and the reason replaced_by is stored.
	if stored.Rotated() {
		if _, rerr := uc.refreshTokenRepository.RevokeAllForUser(ctx, stored.UserID); rerr != nil {
			slog.ErrorContext(ctx, "could not revoke refresh-token family",
				"user_id", stored.UserID.String(), "err", rerr.Error())
		}
		audit.Warn(ctx, audit.RefreshTokenReuse,
			"user_id", stored.UserID.String(),
			"client_ip", clientIP,
			"detail", "rotated refresh token presented again; every session revoked",
		)
		return nil, ErrInvalidRefreshToken
	}

	if stored.Revoked() || stored.Expired(now) {
		return nil, ErrInvalidRefreshToken
	}

	user, err := uc.userRepository.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("could not load user for refresh: %w", err)
	}

	// Re-check account state: this is what turns deactivation from a login-time
	// check into a real one, bounded by the access-token TTL.
	if !user.IsActive() {
		if rerr := uc.refreshTokenRepository.Revoke(ctx, stored.ID); rerr != nil {
			slog.WarnContext(ctx, "could not revoke refresh token for disabled user",
				"user_id", user.ID.String(), "err", rerr.Error())
		}
		audit.Event(ctx, audit.RefreshDenied,
			"user_id", user.ID.String(),
			"client_ip", clientIP,
			"outcome", "disabled",
		)
		return nil, ErrInvalidRefreshToken
	}

	token, newID, err := uc.issuer.issuePair(ctx, user, clientIP, userAgent)
	if err != nil {
		if errors.Is(err, ErrGeneratingToken) {
			return nil, ErrGeneratingToken
		}
		return nil, err
	}

	// Chain the old token to the new one: marks it revoked and records
	// replaced_by, arming reuse detection for the token just handed out.
	if err := uc.refreshTokenRepository.Replace(ctx, stored.ID, newID); err != nil {
		return nil, fmt.Errorf("could not rotate refresh token: %w", err)
	}

	audit.Event(ctx, audit.TokenRefreshed,
		"user_id", user.ID.String(),
		"client_ip", clientIP,
	)
	return token, nil
}
