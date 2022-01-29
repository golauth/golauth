package role

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/domain/entity"
	factoryMock "github.com/golauth/golauth/domain/factory/mock"
	"github.com/golauth/golauth/domain/repository/mock"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type AddRoleSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	rf       *factoryMock.MockRepositoryFactory
	repo     *mock.MockRoleRepository
	addRole  *addRole
}

func TestAddRole(t *testing.T) {
	suite.Run(t, new(AddRoleSuite))
}

func (s *AddRoleSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.rf = factoryMock.NewMockRepositoryFactory(s.mockCtrl)
	s.repo = mock.NewMockRoleRepository(s.mockCtrl)
	s.rf.EXPECT().NewRoleRepository().Return(s.repo)
	s.addRole = NewAddRole(s.rf)
}

func (s *AddRoleSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s AddRoleSuite) TestCreateOk() {
	input := model.RoleRequest{
		Name:        "NEW_ROLE",
		Description: "New Role",
	}
	role := entity.Role{
		ID:           uuid.New(),
		Name:         input.Name,
		Description:  input.Description,
		Enabled:      true,
		CreationDate: time.Now(),
	}
	s.repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&role, nil).Times(1)
	resp, err := s.addRole.Execute(context.Background(), input)
	s.NoError(err)
	s.NotZero(resp)
	s.Equal(role.ID, resp.ID)
	s.Equal(role.Name, resp.Name)
	s.Equal(role.Description, resp.Description)
	s.Equal(role.Enabled, resp.Enabled)
	s.Equal(role.CreationDate, resp.CreationDate)
}

func (s AddRoleSuite) TestCreateNotOk() {
	errMessage := "could not create role"
	input := model.RoleRequest{
		Name:        "NEW_ROLE",
		Description: "New Role",
	}
	s.repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf(errMessage)).Times(1)
	resp, err := s.addRole.Execute(context.Background(), input)
	s.Error(err)
	s.Zero(resp)
	s.EqualError(err, errMessage)
}
