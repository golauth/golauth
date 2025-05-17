package role

import (
	"context"
	"errors"
	"github.com/golauth/golauth/pkg/domain/entity"
	factoryMock "github.com/golauth/golauth/pkg/domain/factory/mock"
	"github.com/golauth/golauth/pkg/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
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

func (s *AddRoleSuite) TestCreateOk() {
	input := entity.Role{
		Name:        "NEW_ROLE",
		Description: "New Role",
		Enabled:     true,
	}
	savedEntity := entity.Role{
		ID:           uuid.New(),
		Name:         "NEW_ROLE",
		Description:  "New Role",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	s.repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&savedEntity, nil).Times(1)
	resp, err := s.addRole.Execute(context.Background(), &input)
	s.NoError(err)
	s.NotZero(resp)
	s.Equal(savedEntity.ID, resp.ID)
	s.Equal(savedEntity.Name, resp.Name)
	s.Equal(savedEntity.Description, resp.Description)
	s.Equal(savedEntity.Enabled, resp.Enabled)
	s.Equal(savedEntity.CreationDate, resp.CreationDate)
}

func (s *AddRoleSuite) TestCreateNotOk() {
	errMessage := "could not create role"
	input := &entity.Role{
		Name:        "NEW_ROLE",
		Description: "New Role",
	}
	s.repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New(errMessage)).Times(1)
	resp, err := s.addRole.Execute(context.Background(), input)
	s.Error(err)
	s.Zero(resp)
	s.EqualError(err, errMessage)
}
