package user

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/src/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AddUserRoleSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	ctx      context.Context

	userRoleRepository *mock.MockUserRoleRepository
	addUserRole        AddUserRole
}

func TestAddUserRole(t *testing.T) {
	suite.Run(t, new(AddUserRoleSuite))
}

func (s *AddUserRoleSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ctx = context.Background()

	s.userRoleRepository = mock.NewMockUserRoleRepository(s.mockCtrl)

	s.addUserRole = NewAddUserRole(s.userRoleRepository)
}

func (s *AddUserRoleSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s AddUserRoleSuite) TestAddUserRoleOK() {
	userId := uuid.New()
	roleId := uuid.New()
	s.userRoleRepository.EXPECT().AddUserRole(s.ctx, userId, roleId).Return(nil).Times(1)
	err := s.addUserRole.Execute(s.ctx, userId, roleId)
	s.NoError(err)
}

func (s AddUserRoleSuite) TestAddUserRoleErr() {
	userId := uuid.New()
	roleId := uuid.New()
	s.userRoleRepository.EXPECT().AddUserRole(s.ctx, userId, roleId).Return(fmt.Errorf("could not add role to user")).Times(1)
	err := s.addUserRole.Execute(s.ctx, userId, roleId)
	s.Error(err)
	s.ErrorAs(fmt.Errorf("could not add role to user"), &err)
}
