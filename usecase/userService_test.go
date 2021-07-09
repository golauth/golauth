package usecase

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"golauth/model"
	"golauth/repository/mock"
	mockSvc "golauth/usecase/mock"
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
	tokenService            *mockSvc.MockTokenService

	svc UserService

	mockUser      model.User
	mockSavedUser model.User
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
	s.tokenService = mockSvc.NewMockTokenService(s.mockCtrl)

	s.svc = NewUserService(s.userRepository, s.roleRepository, s.userRoleRepository, s.userAuthorityRepository, s.tokenService)

	s.mockUser = model.User{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
	}
	s.mockSavedUser = model.User{
		ID:           1,
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
	bcryptDefaultCost = bcrypt.DefaultCost
	s.mockCtrl.Finish()
}

func (s UserServiceSuite) TestCreateUserOk() {
	role := model.Role{ID: 1, Name: "USER", Description: "User", Enabled: true, CreationDate: time.Now()}
	userRole := model.UserRole{UserID: 1, RoleID: 1, CreationDate: time.Now()}
	s.userRepository.EXPECT().Create(gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(defaultRoleName).Return(role, nil).Times(1)
	s.userRoleRepository.EXPECT().AddUserRole(1, 1).Return(userRole, nil).Times(1)

	createUser, err := s.svc.CreateUser(s.mockUser)
	s.NoError(err)
	s.Equal(s.mockSavedUser, createUser)
}

func (s UserServiceSuite) TestCreateUserErrWhenSave() {
	s.userRepository.EXPECT().Create(gomock.Any()).Return(model.User{}, fmt.Errorf("could not create user admin")).Times(1)

	_, err := s.svc.CreateUser(model.User{})
	s.EqualError(err, "could not save user: could not create user admin")
}

func (s UserServiceSuite) TestCreateUserErrFindRole() {
	s.userRepository.EXPECT().Create(gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(defaultRoleName).Return(model.Role{}, fmt.Errorf("could not find role USER")).Times(1)

	_, err := s.svc.CreateUser(s.mockUser)
	s.EqualError(err, "could not fetch default role: could not find role USER")
}

func (s UserServiceSuite) TestCreateUserErrAddUserRole() {
	role := model.Role{ID: 1, Name: "USER", Description: "User", Enabled: true, CreationDate: time.Now()}
	s.userRepository.EXPECT().Create(gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(defaultRoleName).Return(role, nil).Times(1)
	s.userRoleRepository.EXPECT().AddUserRole(1, 1).Return(model.UserRole{}, fmt.Errorf("could not add userrole [user:1:role:1]")).Times(1)

	_, err := s.svc.CreateUser(s.mockUser)
	s.EqualError(err, "could not add default role to user: could not add userrole [user:1:role:1]")
}

func (s UserServiceSuite) TestCreateUserErrGenerateHashPassword() {
	bcryptDefaultCost = 50
	_, err := s.svc.CreateUser(model.User{Password: "1234"})
	s.EqualError(err, "could not generate password: crypto/bcrypt: cost 50 is outside allowed range (4,31)")
}

func (s UserServiceSuite) TestGenerateTokenOk() {
	username := "admin"
	password := "123456"
	encodedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcryptDefaultCost)
	user := model.User{
		ID:           1,
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
	s.userRepository.EXPECT().FindByUsernameWithPassword(username).Return(user, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(user.ID).Return(authorities, nil).Times(1)
	s.tokenService.EXPECT().GenerateJwtToken(user, authorities).Return(token, nil).Times(1)

	tokenResponse, err := s.svc.GenerateToken(username, password)
	s.NoError(err)
	s.NotEmpty(tokenResponse)
	s.Empty(tokenResponse.RefreshToken)
	s.Equal(token, tokenResponse.AccessToken)
}

func (s UserServiceSuite) TestGenerateTokenUserNotFound() {
	username := "admin"
	password := "123456"

	s.userRepository.EXPECT().FindByUsernameWithPassword(username).Return(model.User{}, fmt.Errorf("could not find user by username admin")).Times(1)

	tokenResponse, err := s.svc.GenerateToken(username, password)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Empty(tokenResponse)
}

func (s UserServiceSuite) TestGenerateTokenInvalidPassword() {
	username := "admin"
	password := "123456"
	encodedPassword, _ := bcrypt.GenerateFromPassword([]byte("1234567"), bcryptDefaultCost)
	user := model.User{
		ID:           1,
		Username:     username,
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@ail.com",
		Document:     "1234",
		Password:     string(encodedPassword),
		Enabled:      true,
		CreationDate: time.Now().AddDate(-1, 0, 0),
	}
	s.userRepository.EXPECT().FindByUsernameWithPassword(username).Return(user, nil).Times(1)

	tokenResponse, err := s.svc.GenerateToken(username, password)
	s.ErrorIs(err, ErrInvalidUsernameOrPassword)
	s.Empty(tokenResponse)
}

func (s UserServiceSuite) TestGenerateTokenErrFetchAuthorities() {
	username := "admin"
	password := "123456"
	encodedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcryptDefaultCost)
	user := model.User{
		ID:           1,
		Username:     username,
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@ail.com",
		Document:     "1234",
		Password:     string(encodedPassword),
		Enabled:      true,
		CreationDate: time.Now().AddDate(-1, 0, 0),
	}
	s.userRepository.EXPECT().FindByUsernameWithPassword(username).Return(user, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(user.ID).Return([]string{}, fmt.Errorf("could not find authorities by user admin")).Times(1)

	tokenResponse, err := s.svc.GenerateToken(username, password)
	s.Error(err)
	s.Equal(err.Error(), "error when fetch authorities: could not find authorities by user admin")
	s.Empty(tokenResponse)
}

func (s UserServiceSuite) TestGenerateTokenErrGeneratingToken() {
	username := "admin"
	password := "123456"
	encodedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcryptDefaultCost)
	user := model.User{
		ID:           1,
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
	s.userRepository.EXPECT().FindByUsernameWithPassword(username).Return(user, nil).Times(1)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(user.ID).Return(authorities, nil).Times(1)
	s.tokenService.EXPECT().GenerateJwtToken(user, authorities).Return(model.TokenResponse{}, fmt.Errorf("could not generate token")).Times(1)

	tokenResponse, err := s.svc.GenerateToken(username, password)
	s.ErrorIs(err, ErrGeneratingToken)
	s.Empty(tokenResponse)
}
