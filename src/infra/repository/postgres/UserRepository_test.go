package postgres

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/domain/repository"
	"github.com/golauth/golauth/src/infra/database"
	"github.com/golauth/golauth/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type UserRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       database.Database

	repo repository.UserRepository
}

func TestUserRepository(t *testing.T) {
	ctxContainer, err := tests.ContainerDBStart("./../../../..")
	assert.NoError(t, err)
	s := new(UserRepositorySuite)
	suite.Run(t, s)
	tests.ContainerDBStop(ctxContainer)
}

func (s *UserRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.db = database.NewPGDatabase()

	s.repo = NewUserRepository(s.db)
}

func (s *UserRepositorySuite) TearDownTest() {
	s.db.Close()
	s.mockCtrl.Finish()
}

func (s *UserRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	err := tests.DatasetTest(s.db, "./../../../..", cleanScript, scripts...)
	s.NoError(err)
}

func (s *UserRepositorySuite) TestFindUserWithoutPassword() {
	s.prepareDatabase(true, "add-users.sql")
	id, _ := uuid.Parse("8c61f220-8bb8-48b9-b225-d54dfa6503db")
	u, err := s.repo.FindByID(context.Background(), id)
	s.NoError(err)
	s.NotNil(u)
	s.Equal("admin", u.Username)
	s.Empty(u.Password)
}

func (s *UserRepositorySuite) TestFindUserWithPassword() {
	s.prepareDatabase(true, "add-users.sql")
	u, err := s.repo.FindByUsername(context.Background(), "admin")
	s.NoError(err)
	s.NotNil(u)
	s.Equal("admin", u.Username)
	s.NotEmpty(u.Password)
}

func (s *UserRepositorySuite) TestFindUserByIdWithoutPassword() {
	s.prepareDatabase(true, "add-users.sql")
	userId, _ := uuid.Parse("8c61f220-8bb8-48b9-b225-d54dfa6503db")
	u, err := s.repo.FindByID(context.Background(), userId)
	s.NoError(err)
	s.NotNil(u)
	s.Equal("admin", u.Username)
	s.Zero(u.Password)
}

func (s *UserRepositorySuite) TestCreateNewUserOk() {
	s.prepareDatabase(true, "add-users.sql")
	u := &entity.User{
		Username:     "guest",
		FirstName:    "Guest",
		LastName:     "None",
		Email:        "guest@none.com",
		Document:     "123456",
		Password:     "e10adc3949ba59abbe56e057f20f883e",
		Enabled:      true,
		CreationDate: time.Now(),
	}

	user, err := s.repo.Create(context.Background(), u)
	s.NoError(err)
	s.NotEmpty(user.ID)
}
