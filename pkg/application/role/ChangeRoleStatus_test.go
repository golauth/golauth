package role

import (
	"context"
	"errors"
	"fmt"
	"github.com/golauth/golauth/pkg/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

type ChangeRoleStatusSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl         *gomock.Controller
	ctx              context.Context
	repo             *mock.MockRoleRepository
	changeRoleStatus ChangeRoleStatus
}

func TestChangeRoleStatus(t *testing.T) {
	suite.Run(t, new(ChangeRoleStatusSuite))
}

func (s *ChangeRoleStatusSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.repo = mock.NewMockRoleRepository(s.mockCtrl)
	s.changeRoleStatus = NewChangeRoleStatus(s.repo)
}

func (s *ChangeRoleStatusSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ChangeRoleStatusSuite) TestChangeStatusOk() {
	roleId := uuid.New()
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(true, nil).Times(1)
	s.repo.EXPECT().ChangeStatus(s.ctx, roleId, false).Return(nil).Times(1)
	err := s.changeRoleStatus.Execute(s.ctx, roleId, false)
	s.NoError(err)
}

func (s *ChangeRoleStatusSuite) TestChangeStatusIdNotExists() {
	roleId := uuid.New()
	errMessage := fmt.Sprintf("role with id %s does not exists", roleId)
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(false, nil).Times(1)
	err := s.changeRoleStatus.Execute(s.ctx, roleId, false)
	s.Error(err)
	s.EqualError(err, errMessage)
}

func (s *ChangeRoleStatusSuite) TestChangeStatusExistsErr() {
	errMessage := "could not check if id exists"
	roleId := uuid.New()
	s.repo.EXPECT().ExistsById(s.ctx, roleId).Return(false, errors.New(errMessage)).Times(1)
	err := s.changeRoleStatus.Execute(s.ctx, roleId, false)
	s.Error(err)
	s.EqualError(err, errMessage)
}
