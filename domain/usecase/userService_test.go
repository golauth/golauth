package usecase

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/repository/mock"
	tkSvc "github.com/golauth/golauth/domain/usecase/token/mock"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

type UserServiceSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	userRepository          *mock.MockUserRepository
	roleRepository          *mock.MockRoleRepository
	userRoleRepository      *mock.MockUserRoleRepository
	userAuthorityRepository *mock.MockUserAuthorityRepository
	tokenService            *tkSvc.MockUseCase

	ctx context.Context
	svc UserService

	mockUser      model.UserRequest
	mockSavedUser entity.User
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func (s *UserServiceSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.userRepository = mock.NewMockUserRepository(s.mockCtrl)
	s.roleRepository = mock.NewMockRoleRepository(s.mockCtrl)
	s.userRoleRepository = mock.NewMockUserRoleRepository(s.mockCtrl)
	s.userAuthorityRepository = mock.NewMockUserAuthorityRepository(s.mockCtrl)
	s.tokenService = tkSvc.NewMockUseCase(s.mockCtrl)

	s.ctx = context.Background()
	s.svc = NewUserService(s.userRepository, s.roleRepository, s.userRoleRepository, s.userAuthorityRepository, s.tokenService)

	s.mockUser = model.UserRequest{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
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

func (s *UserServiceSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s UserServiceSuite) TestGenerateTokenOk() {
	username := "admin"
	password := "123456"
	encodedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := entity.User{
		ID:           uuid.New(),
		Username:     username,
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@ail.com",
		Document:     "1234",
		Password:     string(encodedPassword),
		Enabled:      true,
		CreationDate: time.Now().AddDate(-1, 0, 0),
	}
	authorities := []string{"PANEL_EDIT", "PANEL_READ"}
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"
	s.userRepository.EXPECT().FindByUsername(s.ctx, username).Return(user, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(s.ctx, user.ID).Return(authorities, nil).Times(1)
	s.tokenService.EXPECT().GenerateJwtToken(user, authorities).Return(token, nil).Times(1)

	tokenResponse, err := s.svc.GenerateToken(s.ctx, username, password)
	s.NoError(err)
	s.NotEmpty(tokenResponse)
	s.Zero(tokenResponse.RefreshToken)
	s.Equal(token, tokenResponse.AccessToken)
}

func (s UserServiceSuite) TestGenerateTokenUserNotFound() {
	username := "admin"
	password := "123456"

	s.userRepository.EXPECT().FindByUsername(s.ctx, username).Return(entity.User{}, fmt.Errorf("could not find user by username admin")).Times(1)

	tokenResponse, err := s.svc.GenerateToken(s.ctx, username, password)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Empty(tokenResponse)
}

func (s UserServiceSuite) TestGenerateTokenInvalidPassword() {
	username := "admin"
	password := "123456"
	encodedPassword, _ := bcrypt.GenerateFromPassword([]byte("1234567"), bcrypt.DefaultCost)
	user := entity.User{
		ID:           uuid.New(),
		Username:     username,
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@ail.com",
		Document:     "1234",
		Password:     string(encodedPassword),
		Enabled:      true,
		CreationDate: time.Now().AddDate(-1, 0, 0),
	}
	s.userRepository.EXPECT().FindByUsername(s.ctx, username).Return(user, nil).Times(1)

	tokenResponse, err := s.svc.GenerateToken(s.ctx, username, password)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Empty(tokenResponse)
}

func (s UserServiceSuite) TestGenerateTokenErrFetchAuthorities() {
	username := "admin"
	password := "123456"
	encodedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := entity.User{
		ID:           uuid.New(),
		Username:     username,
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@ail.com",
		Document:     "1234",
		Password:     string(encodedPassword),
		Enabled:      true,
		CreationDate: time.Now().AddDate(-1, 0, 0),
	}
	s.userRepository.EXPECT().FindByUsername(s.ctx, username).Return(user, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(s.ctx, user.ID).Return([]string{}, fmt.Errorf("could not find authorities by user admin")).Times(1)

	tokenResponse, err := s.svc.GenerateToken(s.ctx, username, password)
	s.Error(err)
	s.Equal(err.Error(), "error when fetch authorities: could not find authorities by user admin")
	s.Empty(tokenResponse)
}

func (s UserServiceSuite) TestGenerateTokenErrGeneratingToken() {
	username := "admin"
	password := "123456"
	encodedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := entity.User{
		ID:           uuid.New(),
		Username:     username,
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@ail.com",
		Document:     "1234",
		Password:     string(encodedPassword),
		Enabled:      true,
		CreationDate: time.Now().AddDate(-1, 0, 0),
	}
	authorities := []string{"PANEL_EDIT", "PANEL_READ"}
	s.userRepository.EXPECT().FindByUsername(s.ctx, username).Return(user, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(s.ctx, user.ID).Return(authorities, nil).Times(1)
	s.tokenService.EXPECT().GenerateJwtToken(user, authorities).Return("", fmt.Errorf("could not generate token")).Times(1)

	tokenResponse, err := s.svc.GenerateToken(s.ctx, username, password)
	s.ErrorIs(err, ErrGeneratingToken)
	s.Empty(tokenResponse)
}
