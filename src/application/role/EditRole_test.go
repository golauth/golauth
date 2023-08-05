package role

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

type EditRoleSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	ctx      context.Context
	repo     *mock.MockRoleRepository
	editRole EditRole
}

func TestEditRole(t *testing.T) {
	suite.Run(t, new(EditRoleSuite))
}

func (s *EditRoleSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.repo = mock.NewMockRoleRepository(s.mockCtrl)
	s.editRole = NewEditRole(s.repo)
}

func (s *EditRoleSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *EditRoleSuite) TestEditOk() {
	roleId := uuid.New()
	input := &entity.Role{
		ID:          roleId,
		Name:        "NEW_ROLE",
		Description: "Edited Role",
	}
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(true, nil).Times(1)
	s.repo.EXPECT().Edit(s.ctx, input).Return(nil).Times(1)
	err := s.editRole.Execute(s.ctx, roleId, input)
	s.NoError(err)
}

func (s *EditRoleSuite) TestEditIDNotExists() {
	roleId := uuid.New()
	errMessage := fmt.Sprintf("role with id %s does not exists", roleId)
	input := &entity.Role{
		ID:          roleId,
		Name:        "NEW_ROLE",
		Description: "Edited Role",
	}
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(false, nil).Times(1)
	err := s.editRole.Execute(s.ctx, roleId, input)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s *EditRoleSuite) TestEditExistsErr() {
	errMessage := "could not check if id exists"
	roleId := uuid.New()
	input := &entity.Role{
		ID:          roleId,
		Name:        "NEW_ROLE",
		Description: "Edited Role",
	}
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(false, fmt.Errorf(errMessage)).Times(1)
	err := s.editRole.Execute(s.ctx, roleId, input)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s *EditRoleSuite) TestEditErrIdNotMatch() {
	roleId := uuid.New()
	pathId := uuid.New()
	errMessage := fmt.Sprintf("path id[%s] and object_id[%s] does not match", pathId, roleId)
	input := &entity.Role{
		ID:          roleId,
		Name:        "NEW_ROLE",
		Description: "Edited Role",
	}
	s.repo.EXPECT().ExistsById(s.ctx, pathId).Return(true, nil).Times(1)
	err := s.editRole.Execute(s.ctx, pathId, input)
	s.Error(err)
	s.EqualError(err, errMessage)
}
