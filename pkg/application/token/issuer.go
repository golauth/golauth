package token

import (
	"context"
	"fmt"
	"time"

	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/google/uuid"
)

// tokenIssuer mints an access-token / refresh-token pair for a user that has
// already been authenticated and checked as enabled. It is shared by the login
// and refresh use cases so rotation logic exists in exactly one place.
type tokenIssuer struct {
	refreshTokenRepository  repository.RefreshTokenRepository
	userAuthorityRepository repository.UserAuthorityRepository
	jwtToken                GenerateJwtToken
	cfg                     Config
}

// issuePair returns the token pair to hand to the client and the id of the
// newly stored refresh-token row (needed by the caller to chain rotation).
// clientIP and userAgent are recorded against the refresh token so an operator
// can tell sessions apart and spot a stolen one.
func (i tokenIssuer) issuePair(ctx context.Context, user *entity.User, clientIP, userAgent string) (*entity.Token, uuid.UUID, error) {
	// FindAuthoritiesByUserID already filters disabled roles and authorities, so
	// an all-disabled user comes back with an empty list; generateJwtToken then
	// omits the "authorities" claim and every RequireAuthority check denies.
	authorities, err := i.userAuthorityRepository.FindAuthoritiesByUserID(ctx, user.ID)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("error when fetch authorities: %w", err)
	}

	accessToken, err := i.jwtToken.Execute(user, authorities)
	if err != nil {
		return nil, uuid.Nil, ErrGeneratingToken
	}

	plain, hash, err := newOpaqueToken()
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("could not generate refresh token: %w", err)
	}
	now := time.Now()
	created, err := i.refreshTokenRepository.Create(ctx, &entity.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		IssuedAt:  now,
		ExpiresAt: now.Add(i.cfg.RefreshTokenTTL),
		UserAgent: userAgent,
		ClientIP:  clientIP,
	})
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("could not persist refresh token: %w", err)
	}

	return &entity.Token{
		AccessToken:  accessToken,
		RefreshToken: plain,
		ExpiresIn:    int(i.cfg.AccessTokenTTL.Seconds()),
	}, created.ID, nil
}
