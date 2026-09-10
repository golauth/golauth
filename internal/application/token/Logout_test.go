package token

import (
	"context"
	"fmt"
	"testing"

	"github.com/golauth/golauth/internal/domain/entity"
	factoryMock "github.com/golauth/golauth/internal/domain/factory/mock"
	repoMock "github.com/golauth/golauth/internal/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newLogoutForTest(t *testing.T) (Logout, *repoMock.MockRefreshTokenRepository, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	repo := repoMock.NewMockRefreshTokenRepository(ctrl)
	rf := factoryMock.NewMockRepositoryFactory(ctrl)
	rf.EXPECT().NewRefreshTokenRepository().AnyTimes().Return(repo)
	return NewLogout(rf), repo, ctrl
}

func TestLogoutSessionRevokesPresentedToken(t *testing.T) {
	uc, repo, ctrl := newLogoutForTest(t)
	defer ctrl.Finish()
	ctx := context.Background()

	id := uuid.New()
	repo.EXPECT().FindByHash(ctx, hashOpaqueToken("rt")).
		Return(&entity.RefreshToken{ID: id}, nil)
	repo.EXPECT().Revoke(ctx, id).Return(nil)

	require.NoError(t, uc.Session(ctx, "rt"))
}

// An unknown token is a no-op, not an error: logout is idempotent.
func TestLogoutSessionUnknownTokenIsNoOp(t *testing.T) {
	uc, repo, ctrl := newLogoutForTest(t)
	defer ctrl.Finish()
	ctx := context.Background()

	repo.EXPECT().FindByHash(ctx, gomock.Any()).Return(nil, nil)

	require.NoError(t, uc.Session(ctx, "gone"))
}

func TestLogoutSessionPropagatesLookupError(t *testing.T) {
	uc, repo, ctrl := newLogoutForTest(t)
	defer ctrl.Finish()
	ctx := context.Background()

	repo.EXPECT().FindByHash(ctx, gomock.Any()).Return(nil, fmt.Errorf("db down"))

	require.Error(t, uc.Session(ctx, "rt"))
}

func TestLogoutAllSessionsRevokesEveryToken(t *testing.T) {
	uc, repo, ctrl := newLogoutForTest(t)
	defer ctrl.Finish()
	ctx := context.Background()

	userID := uuid.New()
	repo.EXPECT().RevokeAllForUser(ctx, userID).Return(int64(4), nil)

	require.NoError(t, uc.AllSessions(ctx, userID))
}
