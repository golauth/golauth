package usecase

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/entity"
	"golauth/infrastructure/repository/mock"
	"golauth/model"
	"testing"
	"time"
)

type RoleServiceSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	repo *mock.MockRoleRepository

	svc RoleService
}

func TestRoleService(t *testing.T) {
	suite.Run(t, new(RoleServiceSuite))
}

func (s *RoleServiceSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.repo = mock.NewMockRoleRepository(s.mockCtrl)

	s.svc = NewRoleService(s.repo)
}

func (s *RoleServiceSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s RoleServiceSuite) TestCreateOk() {
	req := model.RoleRequest{
		Name:        "NEW_ROLE",
		Description: "New Role",
	}
	role := entity.Role{
		ID:           uuid.New(),
		Name:         req.Name,
		Description:  req.Description,
		Enabled:      true,
		CreationDate: time.Now(),
	}
	s.repo.EXPECT().Create(gomock.Any()).Return(role, nil).Times(1)
	resp, err := s.svc.Create(req)
	s.NoError(err)
	s.NotZero(resp)
	s.Equal(role.ID, resp.ID)
	s.Equal(role.Name, resp.Name)
	s.Equal(role.Description, resp.Description)
	s.Equal(role.Enabled, resp.Enabled)
	s.Equal(role.CreationDate, resp.CreationDate)
}

func (s RoleServiceSuite) TestCreateNotOk() {
	errMessage := "could not create role"
	req := model.RoleRequest{
		Name:        "NEW_ROLE",
		Description: "New Role",
	}
	s.repo.EXPECT().Create(gomock.Any()).Return(entity.Role{}, fmt.Errorf(errMessage)).Times(1)
	resp, err := s.svc.Create(req)
	s.Error(err)
	s.Zero(resp)
	s.EqualError(err, errMessage)
}

func (s RoleServiceSuite) TestEditOk() {
	roleId := uuid.New()
	req := model.RoleRequest{
		ID:          roleId,
		Name:        "NEW_ROLE",
		Description: "Edited Role",
	}
	role := entity.Role{
		ID:          roleId,
		Name:        req.Name,
		Description: req.Description,
	}
	s.repo.EXPECT().ExistsById(roleId).Return(true, nil).Times(1)
	s.repo.EXPECT().Edit(role).Return(nil).Times(1)
	err := s.svc.Edit(roleId, req)
	s.NoError(err)
}

func (s RoleServiceSuite) TestEditIDNotExists() {
	roleId := uuid.New()
	errMessage := fmt.Sprintf("role with id %s does not exists", roleId)
	req := model.RoleRequest{
		ID:          roleId,
		Name:        "NEW_ROLE",
		Description: "Edited Role",
	}
	s.repo.EXPECT().ExistsById(roleId).Return(false, nil).Times(1)
	err := s.svc.Edit(roleId, req)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s RoleServiceSuite) TestEditExistsErr() {
	errMessage := "could not check if id exists"
	roleId := uuid.New()
	req := model.RoleRequest{
		ID:          roleId,
		Name:        "NEW_ROLE",
		Description: "Edited Role",
	}
	s.repo.EXPECT().ExistsById(roleId).Return(false, fmt.Errorf(errMessage)).Times(1)
	err := s.svc.Edit(roleId, req)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s RoleServiceSuite) TestEditErrIdNotMatch() {
	roleId := uuid.New()
	pathId := uuid.New()
	errMessage := fmt.Sprintf("path id[%s] and object_id[%s] does not match", pathId, roleId)
	req := model.RoleRequest{
		ID:          roleId,
		Name:        "NEW_ROLE",
		Description: "Edited Role",
	}
	s.repo.EXPECT().ExistsById(pathId).Return(true, nil).Times(1)
	err := s.svc.Edit(pathId, req)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s RoleServiceSuite) TestChangeStatusOk() {
	roleId := uuid.New()
	s.repo.EXPECT().ExistsById(roleId).Return(true, nil).Times(1)
	s.repo.EXPECT().ChangeStatus(roleId, false).Return(nil).Times(1)
	err := s.svc.ChangeStatus(roleId, false)
	s.NoError(err)
}

func (s RoleServiceSuite) TestChangeStatusIdNotExists() {
	roleId := uuid.New()
	errMessage := fmt.Sprintf("role with id %s does not exists", roleId)
	s.repo.EXPECT().ExistsById(roleId).Return(false, nil).Times(1)
	err := s.svc.ChangeStatus(roleId, false)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s RoleServiceSuite) TestChangeStatusExistsErr() {
	errMessage := "could not check if id exists"
	roleId := uuid.New()
	s.repo.EXPECT().ExistsById(roleId).Return(false, errors.New(errMessage)).Times(1)
	err := s.svc.ChangeStatus(roleId, false)
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
	s.repo.EXPECT().FindByName(role.Name).Return(role, nil).Times(1)
	resp, err := s.svc.FindByName(role.Name)
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
	s.repo.EXPECT().FindByName("ROLE").Return(entity.Role{}, fmt.Errorf(errMessage)).Times(1)
	resp, err := s.svc.FindByName(role.Name)
	s.Error(err)
	s.EqualError(err, errMessage)
	s.Zero(resp)
}
