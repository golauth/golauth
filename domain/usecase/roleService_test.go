package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type RoleServiceSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	ctx  context.Context
	repo *mock.MockRoleRepository

	svc RoleService
}

func TestRoleService(t *testing.T) {
	suite.Run(t, new(RoleServiceSuite))
}

func (s *RoleServiceSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.ctx = context.Background()
	s.repo = mock.NewMockRoleRepository(s.mockCtrl)

	s.svc = NewRoleService(s.repo)
}

func (s *RoleServiceSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s RoleServiceSuite) TestChangeStatusOk() {
	roleId := uuid.New()
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(true, nil).Times(1)
	s.repo.EXPECT().ChangeStatus(s.ctx, roleId, false).Return(nil).Times(1)
	err := s.svc.ChangeStatus(s.ctx, roleId, false)
	s.NoError(err)
}

func (s RoleServiceSuite) TestChangeStatusIdNotExists() {
	roleId := uuid.New()
	errMessage := fmt.Sprintf("role with id %s does not exists", roleId)
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(false, nil).Times(1)
	err := s.svc.ChangeStatus(s.ctx, roleId, false)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s RoleServiceSuite) TestChangeStatusExistsErr() {
	errMessage := "could not check if id exists"
	roleId := uuid.New()
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(false, errors.New(errMessage)).Times(1)
	err := s.svc.ChangeStatus(s.ctx, roleId, false)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s RoleServiceSuite) TestFindByNameOk() {
	roleId := uuid.New()
	role := entity.Role{
		ID:           roleId,
		Name:         "ROLE",
		Description:  "Role description",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	s.repo.EXPECT().FindByName(s.ctx, role.Name).Return(&role, nil).Times(1)
	resp, err := s.svc.FindByName(s.ctx, role.Name)
	s.NoError(err)
	s.NotZero(resp)
	s.Equal(role.ID, resp.ID)
	s.Equal(role.Name, resp.Name)
	s.Equal(role.Description, resp.Description)
	s.Equal(role.Enabled, resp.Enabled)
	s.Equal(role.CreationDate, resp.CreationDate)
}

func (s RoleServiceSuite) TestFindByNameNotOk() {
	roleId := uuid.New()
	role := entity.Role{
		ID:           roleId,
		Name:         "ROLE",
		Description:  "Role description",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	errMessage := "could not find role by name"
	s.repo.EXPECT().FindByName(s.ctx, "ROLE").Return(nil, fmt.Errorf(errMessage)).Times(1)
	resp, err := s.svc.FindByName(s.ctx, role.Name)
	s.Error(err)
	s.EqualError(err, errMessage)
	s.Zero(resp)
}
