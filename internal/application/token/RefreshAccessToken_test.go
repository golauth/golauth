package token

import (
	"context"
	"fmt"
	"testing"
	"time"

	tokenMock "github.com/golauth/golauth/internal/application/token/mock"
	"github.com/golauth/golauth/internal/domain/entity"
	factoryMock "github.com/golauth/golauth/internal/domain/factory/mock"
	repoMock "github.com/golauth/golauth/internal/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type RefreshAccessTokenSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	userRepository          *repoMock.MockUserRepository
	userAuthorityRepository *repoMock.MockUserAuthorityRepository
	refreshTokenRepository  *repoMock.MockRefreshTokenRepository
	jwtToken                *tokenMock.MockGenerateJwtToken

	ctx     context.Context
	uc      RefreshAccessToken
	user    *entity.User
	present string
	stored  *entity.RefreshToken
}

func TestRefreshAccessToken(t *testing.T) {
	suite.Run(t, new(RefreshAccessTokenSuite))
}

func (s *RefreshAccessTokenSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.userRepository = repoMock.NewMockUserRepository(s.mockCtrl)
	s.userAuthorityRepository = repoMock.NewMockUserAuthorityRepository(s.mockCtrl)
	s.refreshTokenRepository = repoMock.NewMockRefreshTokenRepository(s.mockCtrl)
	s.jwtToken = tokenMock.NewMockGenerateJwtToken(s.mockCtrl)

	rf := factoryMock.NewMockRepositoryFactory(s.mockCtrl)
	rf.EXPECT().NewRefreshTokenRepository().AnyTimes().Return(s.refreshTokenRepository)
	rf.EXPECT().NewUserRepository().AnyTimes().Return(s.userRepository)
	rf.EXPECT().NewUserAuthorityRepository().AnyTimes().Return(s.userAuthorityRepository)

	s.ctx = context.Background()
	s.uc = NewRefreshAccessToken(rf, s.jwtToken, DefaultConfig())

	s.user = &entity.User{ID: uuid.New(), Username: "admin", Enabled: true}
	s.present = "presented-refresh-token"
	s.stored = &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    s.user.ID,
		TokenHash: hashOpaqueToken(s.present),
		IssuedAt:  time.Now().Add(-time.Hour),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func (s *RefreshAccessTokenSuite) TearDownTest() { s.mockCtrl.Finish() }

// A valid refresh token yields a new pair and rotates (revokes + chains) the
// presented one.
func (s *RefreshAccessTokenSuite) TestRefreshReturnsNewPairAndRotatesOld() {
	newRowID := uuid.New()
	s.refreshTokenRepository.EXPECT().FindByHash(s.ctx, s.stored.TokenHash).Return(s.stored, nil)
	s.userRepository.EXPECT().FindByID(s.ctx, s.user.ID).Return(s.user, nil)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(s.ctx, s.user.ID).Return([]string{"USER"}, nil)
	s.jwtToken.EXPECT().Execute(s.user, []string{"USER"}).Return("new.access.jwt", nil)
	s.refreshTokenRepository.EXPECT().Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, rt *entity.RefreshToken) (*entity.RefreshToken, error) {
			s.Equal(s.user.ID, rt.UserID)
			s.NotEqual(s.stored.TokenHash, rt.TokenHash)
			rt.ID = newRowID
			return rt, nil
		})
	s.refreshTokenRepository.EXPECT().Replace(s.ctx, s.stored.ID, newRowID).Return(nil)

	out, err := s.uc.Execute(s.ctx, s.present, "203.0.113.9", "curl/8")
	s.NoError(err)
	s.Equal("new.access.jwt", out.AccessToken)
	s.NotEmpty(out.RefreshToken)
	s.NotEqual(s.present, out.RefreshToken)
	s.Equal(int(DefaultAccessTokenTTL.Seconds()), out.ExpiresIn)
}

func (s *RefreshAccessTokenSuite) TestUnknownTokenIsInvalid() {
	s.refreshTokenRepository.EXPECT().FindByHash(s.ctx, gomock.Any()).Return(nil, nil)

	out, err := s.uc.Execute(s.ctx, "nope", "", "")
	s.ErrorIs(err, ErrInvalidRefreshToken)
	s.Nil(out)
}

// Reuse of an already-rotated token revokes the whole family and fails.
func (s *RefreshAccessTokenSuite) TestReuseOfRotatedTokenRevokesFamily() {
	replacedBy := uuid.New()
	s.stored.ReplacedBy = &replacedBy
	s.refreshTokenRepository.EXPECT().FindByHash(s.ctx, s.stored.TokenHash).Return(s.stored, nil)
	s.refreshTokenRepository.EXPECT().RevokeAllForUser(s.ctx, s.user.ID).Return(int64(3), nil)

	out, err := s.uc.Execute(s.ctx, s.present, "", "")
	s.ErrorIs(err, ErrInvalidRefreshToken)
	s.Nil(out)
}

func (s *RefreshAccessTokenSuite) TestRevokedTokenIsInvalid() {
	revokedAt := time.Now().Add(-time.Minute)
	s.stored.RevokedAt = &revokedAt
	s.refreshTokenRepository.EXPECT().FindByHash(s.ctx, s.stored.TokenHash).Return(s.stored, nil)

	out, err := s.uc.Execute(s.ctx, s.present, "", "")
	s.ErrorIs(err, ErrInvalidRefreshToken)
	s.Nil(out)
}

func (s *RefreshAccessTokenSuite) TestExpiredTokenIsInvalid() {
	s.stored.ExpiresAt = time.Now().Add(-time.Second)
	s.refreshTokenRepository.EXPECT().FindByHash(s.ctx, s.stored.TokenHash).Return(s.stored, nil)

	out, err := s.uc.Execute(s.ctx, s.present, "", "")
	s.ErrorIs(err, ErrInvalidRefreshToken)
	s.Nil(out)
}

// A disabled user cannot refresh; the presented token is revoked on the way out.
func (s *RefreshAccessTokenSuite) TestDisabledUserCannotRefresh() {
	s.user.Enabled = false
	s.refreshTokenRepository.EXPECT().FindByHash(s.ctx, s.stored.TokenHash).Return(s.stored, nil)
	s.userRepository.EXPECT().FindByID(s.ctx, s.user.ID).Return(s.user, nil)
	s.refreshTokenRepository.EXPECT().Revoke(s.ctx, s.stored.ID).Return(nil)

	out, err := s.uc.Execute(s.ctx, s.present, "", "")
	s.ErrorIs(err, ErrInvalidRefreshToken)
	s.Nil(out)
}

func (s *RefreshAccessTokenSuite) TestRepositoryErrorIsWrapped() {
	s.refreshTokenRepository.EXPECT().FindByHash(s.ctx, gomock.Any()).
		Return(nil, fmt.Errorf("connection reset"))

	out, err := s.uc.Execute(s.ctx, s.present, "", "")
	s.Error(err)
	s.NotErrorIs(err, ErrInvalidRefreshToken)
	s.Nil(out)
}
