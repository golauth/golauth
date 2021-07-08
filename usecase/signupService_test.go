package usecase

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"golauth/model"
	"golauth/repository/mock"
	"testing"
	"time"
)

type SignupServiceSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	userRepository     *mock.MockUserRepository
	roleRepository     *mock.MockRoleRepository
	userRoleRepository *mock.MockUserRoleRepository

	svc SignupService

	mockUser      model.User
	mockSavedUser model.User
}

func TestSignupService(t *testing.T) {
	suite.Run(t, new(SignupServiceSuite))
}

func (s *SignupServiceSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.userRepository = mock.NewMockUserRepository(s.mockCtrl)
	s.roleRepository = mock.NewMockRoleRepository(s.mockCtrl)
	s.userRoleRepository = mock.NewMockUserRoleRepository(s.mockCtrl)

	s.svc = NewSignupService(s.userRepository, s.roleRepository, s.userRoleRepository)

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

func (s *SignupServiceSuite) TearDownTest() {
	bcryptDefaultCost = bcrypt.DefaultCost
	s.mockCtrl.Finish()
}

func (s SignupServiceSuite) TestCreateUserOk() {
	role := model.Role{ID: 1, Name: "USER", Description: "User", Enabled: true, CreationDate: time.Now()}
	userRole := model.UserRole{UserID: 1, RoleID: 1, CreationDate: time.Now()}
	s.userRepository.EXPECT().Create(gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(defaultRoleName).Return(role, nil).Times(1)
	s.userRoleRepository.EXPECT().AddUserRole(1, 1).Return(userRole, nil).Times(1)

	createUser, err := s.svc.CreateUser(s.mockUser)
	s.NoError(err)
	s.Equal(s.mockSavedUser, createUser)
}

func (s SignupServiceSuite) TestCreateUserErrWhenSave() {
	s.userRepository.EXPECT().Create(gomock.Any()).Return(model.User{}, fmt.Errorf("could not create user admin")).Times(1)

	_, err := s.svc.CreateUser(model.User{})
	s.EqualError(err, "could not save user: could not create user admin")
}

func (s SignupServiceSuite) TestCreateUserErrFindRole() {
	s.userRepository.EXPECT().Create(gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(defaultRoleName).Return(model.Role{}, fmt.Errorf("could not find role USER")).Times(1)

	_, err := s.svc.CreateUser(s.mockUser)
	s.EqualError(err, "could not fetch default role: could not find role USER")
}

func (s SignupServiceSuite) TestCreateUserErrAddUserRole() {
	role := model.Role{ID: 1, Name: "USER", Description: "User", Enabled: true, CreationDate: time.Now()}
	s.userRepository.EXPECT().Create(gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(defaultRoleName).Return(role, nil).Times(1)
	s.userRoleRepository.EXPECT().AddUserRole(1, 1).Return(model.UserRole{}, fmt.Errorf("could not add userrole [user:1:role:1]")).Times(1)

	_, err := s.svc.CreateUser(s.mockUser)
	s.EqualError(err, "could not add default role to user: could not add userrole [user:1:role:1]")
}

func (s SignupServiceSuite) TestCreateUserErrGenerateHashPassword() {
	bcryptDefaultCost = 50
	_, err := s.svc.CreateUser(model.User{Password: "1234"})
	s.EqualError(err, "could not generate password: crypto/bcrypt: cost 50 is outside allowed range (4,31)")
}
