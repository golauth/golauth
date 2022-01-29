package user

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/domain/entity"
	mock2 "github.com/golauth/golauth/domain/factory/mock"
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

type CreateUserSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	repoFactory        *mock2.MockRepositoryFactory
	userRepository     *mock.MockUserRepository
	roleRepository     *mock.MockRoleRepository
	userRoleRepository *mock.MockUserRoleRepository
	tokenService       *tkSvc.MockUseCase

	ctx        context.Context
	createUser CreateUser

	mockUser      model.UserRequest
	mockSavedUser entity.User
}

func TestCreateUser(t *testing.T) {
	suite.Run(t, new(CreateUserSuite))
}

func (s *CreateUserSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.repoFactory = mock2.NewMockRepositoryFactory(s.mockCtrl)
	s.userRepository = mock.NewMockUserRepository(s.mockCtrl)
	s.roleRepository = mock.NewMockRoleRepository(s.mockCtrl)
	s.userRoleRepository = mock.NewMockUserRoleRepository(s.mockCtrl)
	s.tokenService = tkSvc.NewMockUseCase(s.mockCtrl)

	s.repoFactory.EXPECT().NewRoleRepository().AnyTimes().Return(s.roleRepository)
	s.repoFactory.EXPECT().NewUserRoleRepository().AnyTimes().Return(s.userRoleRepository)
	s.repoFactory.EXPECT().NewUserRepository().AnyTimes().Return(s.userRepository)

	s.ctx = context.Background()
	s.createUser = NewCreateUser(s.repoFactory)

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

func (s *CreateUserSuite) TearDownTest() {
	bcryptDefaultCost = bcrypt.DefaultCost
	s.mockCtrl.Finish()
}

func (s CreateUserSuite) TestCreateUserOk() {
	roleId := uuid.New()
	userId := s.mockSavedUser.ID
	role := entity.Role{ID: roleId, Name: "USER", Description: "User", Enabled: true, CreationDate: time.Now()}
	s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(s.ctx, defaultRoleName).Return(&role, nil).Times(1)
	s.userRoleRepository.EXPECT().AddUserRole(s.ctx, userId, roleId).Return(nil).Times(1)

	createUser, err := s.createUser.Execute(s.ctx, s.mockUser)
	s.NoError(err)
	s.Equal(s.mockSavedUser.ID, createUser.ID)
	s.Equal(s.mockSavedUser.Username, createUser.Username)
}

func (s CreateUserSuite) TestCreateUserErrWhenSave() {
	s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(entity.User{}, fmt.Errorf("could not create user admin")).Times(1)

	_, err := s.createUser.Execute(s.ctx, model.UserRequest{})
	s.EqualError(err, "could not save user: could not create user admin")
}

func (s CreateUserSuite) TestCreateUserErrFindRole() {
	s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(s.ctx, defaultRoleName).Return(nil, fmt.Errorf("could not find role USER")).Times(1)

	_, err := s.createUser.Execute(s.ctx, s.mockUser)
	s.EqualError(err, "could not fetch default role: could not find role USER")
}

func (s CreateUserSuite) TestCreateUserErrAddUserRole() {
	roleId := uuid.New()
	userId := s.mockSavedUser.ID
	role := entity.Role{ID: roleId, Name: "USER", Description: "User", Enabled: true, CreationDate: time.Now()}
	s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(s.mockSavedUser, nil).Times(1)
	s.roleRepository.EXPECT().FindByName(s.ctx, defaultRoleName).Return(&role, nil).Times(1)
	s.userRoleRepository.
		EXPECT().
		AddUserRole(s.ctx, userId, roleId).Return(fmt.Errorf("could not add userrole [user:%s:role:%s]", userId, roleId)).
		Times(1)

	_, err := s.createUser.Execute(s.ctx, s.mockUser)
	s.EqualError(err, fmt.Sprintf("could not add default role to user: could not add userrole [user:%s:role:%s]", userId, roleId))
}

func (s CreateUserSuite) TestCreateUserErrGenerateHashPassword() {
	bcryptDefaultCost = 50
	_, err := s.createUser.Execute(s.ctx, model.UserRequest{Password: "1234"})
	s.EqualError(err, "could not generate password: crypto/bcrypt: cost 50 is outside allowed range (4,31)")
}
