package role

import (
	"context"
	"errors"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type FindRoleByNameSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	ctx      context.Context
	repo     *mock.MockRoleRepository
	finding  FindRoleByName
}

func TestFindRoleByName(t *testing.T) {
	suite.Run(t, new(FindRoleByNameSuite))
}

func (s *FindRoleByNameSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.repo = mock.NewMockRoleRepository(s.mockCtrl)

	s.finding = NewFindRoleByName(s.repo)
}

func (s *FindRoleByNameSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *FindRoleByNameSuite) TestFindByNameOk() {
	roleId := uuid.New()
	role := entity.Role{
		ID:           roleId,
		Name:         "ROLE",
		Description:  "Role description",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	s.repo.EXPECT().FindByName(s.ctx, role.Name).Return(&role, nil).Times(1)
	resp, err := s.finding.Execute(s.ctx, role.Name)
	s.NoError(err)
	s.NotZero(resp)
	s.Equal(role.ID, resp.ID)
	s.Equal(role.Name, resp.Name)
	s.Equal(role.Description, resp.Description)
	s.Equal(role.Enabled, resp.Enabled)
	s.Equal(role.CreationDate, resp.CreationDate)
}

func (s *FindRoleByNameSuite) TestFindByNameNotOk() {
	roleId := uuid.New()
	role := entity.Role{
		ID:           roleId,
		Name:         "ROLE",
		Description:  "Role description",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	errMessage := "could not find role by name"
	s.repo.EXPECT().FindByName(s.ctx, "ROLE").Return(nil, errors.New(errMessage)).Times(1)
	resp, err := s.finding.Execute(s.ctx, role.Name)
	s.Error(err)
	s.EqualError(err, errMessage)
	s.Zero(resp)
}
