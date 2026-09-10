package token

import (
	"context"
	"fmt"
	"testing"
	"time"

	tokenMock "github.com/golauth/golauth/pkg/application/token/mock"
	"github.com/golauth/golauth/pkg/domain/entity"
	factoryMock "github.com/golauth/golauth/pkg/domain/factory/mock"
	repoMock "github.com/golauth/golauth/pkg/domain/repository/mock"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

const testClientIP = "198.51.100.9"

type GenerateTokenSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	userRepository          *repoMock.MockUserRepository
	roleRepository          *repoMock.MockRoleRepository
	userRoleRepository      *repoMock.MockUserRoleRepository
	userAuthorityRepository *repoMock.MockUserAuthorityRepository
	loginAttemptRepository  *repoMock.MockLoginAttemptRepository
	jwtToken                *tokenMock.MockGenerateJwtToken

	repoFactory *factoryMock.MockRepositoryFactory

	ctx           context.Context
	generateToken GenerateToken

	mockUser      model.CreateUserRequest
	mockSavedUser entity.User
}

func TestGenerateToken(t *testing.T) {
	suite.Run(t, new(GenerateTokenSuite))
}

func (s *GenerateTokenSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.userRepository = repoMock.NewMockUserRepository(s.mockCtrl)
	s.roleRepository = repoMock.NewMockRoleRepository(s.mockCtrl)
	s.userRoleRepository = repoMock.NewMockUserRoleRepository(s.mockCtrl)
	s.userAuthorityRepository = repoMock.NewMockUserAuthorityRepository(s.mockCtrl)
	s.loginAttemptRepository = repoMock.NewMockLoginAttemptRepository(s.mockCtrl)
	s.jwtToken = tokenMock.NewMockGenerateJwtToken(s.mockCtrl)
	s.repoFactory = factoryMock.NewMockRepositoryFactory(s.mockCtrl)
	s.repoFactory.EXPECT().NewRoleRepository().AnyTimes().Return(s.roleRepository)
	s.repoFactory.EXPECT().NewUserRoleRepository().AnyTimes().Return(s.userRoleRepository)
	s.repoFactory.EXPECT().NewUserAuthorityRepository().AnyTimes().Return(s.userAuthorityRepository)
	s.repoFactory.EXPECT().NewUserRepository().AnyTimes().Return(s.userRepository)
	s.repoFactory.EXPECT().NewLoginAttemptRepository().AnyTimes().Return(s.loginAttemptRepository)

	s.ctx = context.Background()
	s.generateToken = NewGenerateToken(s.repoFactory, s.jwtToken, DefaultLockoutPolicy)

	s.mockUser = model.CreateUserRequest{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
	}
	s.mockSavedUser = entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@il.com",
		Document:     "1234",
		Password:     "4567",
		Enabled:      true,
		CreationDate: time.Now(),
	}
}

func (s *GenerateTokenSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// userWithPassword returns an enabled user whose stored hash matches plain.
func userWithPassword(plain string) *entity.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return &entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@ail.com",
		Document:     "1234",
		Password:     string(hash),
		Enabled:      true,
		CreationDate: time.Now().AddDate(-1, 0, 0),
	}
}

func (s *GenerateTokenSuite) TestGenerateTokenOk() {
	user := userWithPassword("123456")
	authorities := []string{"PANEL_EDIT", "PANEL_READ"}
	token := "header.payload.signature"
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(user, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, user.ID).Return(nil, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(s.ctx, user.ID).Return(authorities, nil).Times(1)
	s.jwtToken.EXPECT().Execute(user, authorities).Return(token, nil).Times(1)

	tokenResponse, err := s.generateToken.Execute(s.ctx, "admin", "123456", testClientIP)
	s.NoError(err)
	s.NotEmpty(tokenResponse)
	s.Equal(token, tokenResponse.AccessToken)
}

func (s *GenerateTokenSuite) TestGenerateTokenUserNotFound() {
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").
		Return(nil, fmt.Errorf("could not find user by username admin")).Times(1)

	tokenResponse, err := s.generateToken.Execute(s.ctx, "admin", "123456", testClientIP)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Empty(tokenResponse)
}

// TestGenerateTokenUnknownUserPaysBcryptCost asserts the code path, not wall
// time: the unknown-user branch runs the password comparison against the fixed
// dummy hash so its latency matches the wrong-password branch.
func (s *GenerateTokenSuite) TestGenerateTokenUnknownUserPaysBcryptCost() {
	original := comparePassword
	defer func() { comparePassword = original }()
	var gotHash, gotPassword []byte
	calls := 0
	comparePassword = func(hashedPassword, password []byte) error {
		calls++
		gotHash, gotPassword = hashedPassword, password
		return original(hashedPassword, password)
	}

	s.userRepository.EXPECT().FindByUsername(s.ctx, "ghost").
		Return(nil, fmt.Errorf("no such user")).Times(1)

	_, err := s.generateToken.Execute(s.ctx, "ghost", "guess", testClientIP)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Equal(1, calls)
	s.Equal(dummyBcryptHash, string(gotHash))
	s.Equal("guess", string(gotPassword))
}

func (s *GenerateTokenSuite) TestGenerateTokenInvalidPassword() {
	user := userWithPassword("1234567")
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(user, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, user.ID).Return(nil, nil).Times(1)
	// One failure, below the default threshold: recorded, but no lock yet.
	s.loginAttemptRepository.EXPECT().RegisterFailure(s.ctx, user.ID, time.Time{}).Return(nil).Times(1)

	tokenResponse, err := s.generateToken.Execute(s.ctx, "admin", "123456", testClientIP)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Empty(tokenResponse)
}

func (s *GenerateTokenSuite) TestGenerateTokenErrFetchAuthorities() {
	user := userWithPassword("123456")
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(user, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, user.ID).Return(nil, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(s.ctx, user.ID).
		Return([]string{}, fmt.Errorf("could not find authorities by user admin")).Times(1)

	tokenResponse, err := s.generateToken.Execute(s.ctx, "admin", "123456", testClientIP)
	s.Error(err)
	s.Equal("error when fetch authorities: could not find authorities by user admin", err.Error())
	s.Empty(tokenResponse)
}

func (s *GenerateTokenSuite) TestGenerateTokenErrGeneratingToken() {
	user := userWithPassword("123456")
	authorities := []string{"PANEL_EDIT", "PANEL_READ"}
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(user, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, user.ID).Return(nil, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(s.ctx, user.ID).Return(authorities, nil).Times(1)
	s.jwtToken.EXPECT().Execute(user, authorities).Return("", fmt.Errorf("could not generate token")).Times(1)

	tokenResponse, err := s.generateToken.Execute(s.ctx, "admin", "123456", testClientIP)
	s.ErrorIs(err, ErrGeneratingToken)
	s.Empty(tokenResponse)
}

// TestGenerateTokenDisabledUser: a deactivated account with the correct
// password is refused, byte-for-byte like a wrong password.
func (s *GenerateTokenSuite) TestGenerateTokenDisabledUser() {
	user := userWithPassword("123456")
	user.Enabled = false
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(user, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, user.ID).Return(nil, nil).Times(1)

	tokenResponse, err := s.generateToken.Execute(s.ctx, "admin", "123456", testClientIP)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Empty(tokenResponse)

	wrongPwUser := userWithPassword("the-real-one")
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(wrongPwUser, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, wrongPwUser.ID).Return(nil, nil).Times(1)
	s.loginAttemptRepository.EXPECT().RegisterFailure(s.ctx, wrongPwUser.ID, time.Time{}).Return(nil).Times(1)
	wrongResponse, wrongErr := s.generateToken.Execute(s.ctx, "admin", "not-the-password", testClientIP)
	s.Equal(err, wrongErr)
	s.Equal(tokenResponse, wrongResponse)
}

// TestGenerateTokenLocksAfterThreshold: the failure that reaches the threshold
// is stored with a lock window in the future.
func (s *GenerateTokenSuite) TestGenerateTokenLocksAfterThreshold() {
	policy := LockoutPolicy{Threshold: 3, BaseDelay: time.Minute, MaxDelay: time.Hour}
	uc := NewGenerateToken(s.repoFactory, s.jwtToken, policy)

	user := userWithPassword("correct-horse")
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(user, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, user.ID).
		Return(&entity.LoginAttempt{UserID: user.ID, FailedCount: 2}, nil).Times(1)

	var lockedUntil time.Time
	s.loginAttemptRepository.EXPECT().RegisterFailure(s.ctx, user.ID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, until time.Time) error {
			lockedUntil = until
			return nil
		}).Times(1)

	_, err := uc.Execute(s.ctx, "admin", "wrong", testClientIP)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.WithinDuration(time.Now().Add(time.Minute), lockedUntil, 5*time.Second)
}

// TestGenerateTokenRejectsLockedAccount: a live lock short-circuits before any
// password work, so even the correct password fails and nothing new is recorded.
func (s *GenerateTokenSuite) TestGenerateTokenRejectsLockedAccount() {
	original := comparePassword
	defer func() { comparePassword = original }()
	comparePassword = func([]byte, []byte) error {
		s.Fail("password comparison must not run for a locked account")
		return nil
	}

	user := userWithPassword("correct-horse")
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(user, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, user.ID).
		Return(&entity.LoginAttempt{UserID: user.ID, FailedCount: 9, LockedUntil: time.Now().Add(10 * time.Minute)}, nil).Times(1)

	tokenResponse, err := s.generateToken.Execute(s.ctx, "admin", "correct-horse", testClientIP)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Empty(tokenResponse)
}

// TestGenerateTokenResetsCounterOnSuccess: a correct password on an account
// with prior (expired) failures clears the counter.
func (s *GenerateTokenSuite) TestGenerateTokenResetsCounterOnSuccess() {
	user := userWithPassword("correct-horse")
	authorities := []string{"USER"}
	s.userRepository.EXPECT().FindByUsername(s.ctx, "admin").Return(user, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Get(s.ctx, user.ID).
		Return(&entity.LoginAttempt{UserID: user.ID, FailedCount: 2}, nil).Times(1)
	s.loginAttemptRepository.EXPECT().Reset(s.ctx, user.ID).Return(nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(s.ctx, user.ID).Return(authorities, nil).Times(1)
	s.jwtToken.EXPECT().Execute(user, authorities).Return("a.b.c", nil).Times(1)

	tokenResponse, err := s.generateToken.Execute(s.ctx, "admin", "correct-horse", testClientIP)
	s.NoError(err)
	s.Equal("a.b.c", tokenResponse.AccessToken)
}
