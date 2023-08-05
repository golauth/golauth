package user

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/src/domain/entity"
	factoryMock "github.com/golauth/golauth/src/domain/factory/mock"
	repoMock "github.com/golauth/golauth/src/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

type CreateUserSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	repoFactory        *factoryMock.MockRepositoryFactory
	userRepository     *repoMock.MockUserRepository
	roleRepository     *repoMock.MockRoleRepository
	userRoleRepository *repoMock.MockUserRoleRepository

	ctx        context.Context
	createUser CreateUser

	input         *entity.User
	mockSavedUser *entity.User
}

func TestCreateUser(t *testing.T) {
	suite.Run(t, new(CreateUserSuite))
}

func (s *CreateUserSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.repoFactory = factoryMock.NewMockRepositoryFactory(s.mockCtrl)
	s.userRepository = repoMock.NewMockUserRepository(s.mockCtrl)
	s.roleRepository = repoMock.NewMockRoleRepository(s.mockCtrl)
	s.userRoleRepository = repoMock.NewMockUserRoleRepository(s.mockCtrl)

	s.repoFactory.EXPECT().NewRoleRepository().AnyTimes().Return(s.roleRepository)
	s.repoFactory.EXPECT().NewUserRoleRepository().AnyTimes().Return(s.userRoleRepository)
	s.repoFactory.EXPECT().NewUserRepository().AnyTimes().Return(s.userRepository)

	s.ctx = context.Background()
	s.createUser = NewCreateUser(s.repoFactory)

	s.input = &entity.User{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
	}
	s.mockSavedUser = &entity.User{
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

func (s *CreateUserSuite) TearDownTest() {
	bcryptDefaultCost = bcrypt.DefaultCost
	s.mockCtrl.Finish()
}

func (s *CreateUserSuite) TestCreateUserOk() {
	roleId := uuid.New()
	userId := s.mockSavedUser.ID
	role := entity.Role{ID: roleId, Name: "USER", Description: "User", Enabled: true, CreationDate: time.Now()}
	s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(s.ctx, defaultRoleName).Return(&role, nil).Times(1)
	s.userRoleRepository.EXPECT().AddUserRole(s.ctx, userId, roleId).Return(nil).Times(1)

	createUser, err := s.createUser.Execute(s.ctx, s.input)
	s.NoError(err)
	s.Equal(s.mockSavedUser.ID, createUser.ID)
	s.Equal(s.mockSavedUser.Username, createUser.Username)
}

func (s *CreateUserSuite) TestCreateUserErrWhenSave() {
	s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("could not create user admin")).Times(1)

	_, err := s.createUser.Execute(s.ctx, &entity.User{})
	s.EqualError(err, "could not save user: could not create user admin")
}

func (s *CreateUserSuite) TestCreateUserErrFindRole() {
	s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(s.ctx, defaultRoleName).Return(nil, fmt.Errorf("could not find role USER")).Times(1)

	_, err := s.createUser.Execute(s.ctx, s.input)
	s.EqualError(err, "could not fetch default role: could not find role USER")
}

func (s *CreateUserSuite) TestCreateUserErrAddUserRole() {
	roleId := uuid.New()
	userId := s.mockSavedUser.ID
	role := entity.Role{ID: roleId, Name: "USER", Description: "User", Enabled: true, CreationDate: time.Now()}
	s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(s.ctx, defaultRoleName).Return(&role, nil).Times(1)
	s.userRoleRepository.
		EXPECT().
		AddUserRole(s.ctx, userId, roleId).Return(fmt.Errorf("could not add userrole [user:%s:role:%s]", userId, roleId)).
		Times(1)

	_, err := s.createUser.Execute(s.ctx, s.input)
	s.EqualError(err, fmt.Sprintf("could not add default role to user: could not add userrole [user:%s:role:%s]", userId, roleId))
}

func (s *CreateUserSuite) TestCreateUserErrGenerateHashPassword() {
	bcryptDefaultCost = 50
	_, err := s.createUser.Execute(s.ctx, &entity.User{Password: "1234"})
	s.EqualError(err, "could not generate password: crypto/bcrypt: cost 50 is outside allowed range (4,31)")
}
