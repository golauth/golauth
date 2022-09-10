package postgres

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/domain/repository"
	"github.com/golauth/golauth/src/infra/database"
	"github.com/golauth/golauth/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UserRoleRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       database.Database

	repo repository.UserRoleRepository
}

func TestUserRoleRepository(t *testing.T) {
	ctxContainer, err := tests.ContainerDBStart("./../../../..")
	assert.NoError(t, err)
	s := new(UserRoleRepositorySuite)
	suite.Run(t, s)
	tests.ContainerDBStop(ctxContainer)
}

func (s *UserRoleRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.db = database.NewPGDatabase()

	s.repo = NewUserRoleRepository(s.db)
}

func (s *UserRoleRepositorySuite) TearDownTest() {
	s.db.Close()
	s.mockCtrl.Finish()
}

func (s *UserRoleRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	err := tests.DatasetTest(s.db, "./../../../..", cleanScript, scripts...)
	s.NoError(err)
}

func (s *UserRoleRepositorySuite) TestAddUserRole() {
	u := &entity.User{
		Username:  "guest",
		FirstName: "Guest",
		LastName:  "None",
		Email:     "guest@none.com",
		Document:  "123456",
		Password:  "e10adc3949ba59abbe56e057f20f883e",
		Enabled:   true,
	}
	user, err := NewUserRepository(s.db).Create(context.Background(), u)
	s.NoError(err)
	s.NotNil(user)

	role, err := NewRoleRepository(s.db).FindByName(context.Background(), "USER")
	s.NoError(err)
	s.NotNil(role)

	err = s.repo.AddUserRole(context.Background(), user.ID, role.ID)
	s.NoError(err)
}
