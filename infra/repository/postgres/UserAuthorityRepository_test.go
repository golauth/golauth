package postgres

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/database"
	"github.com/golauth/golauth/ops"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UserAuthorityRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       database.Database
	repo     repository.UserAuthorityRepository

	userAdminId uuid.UUID
}

func TestUserAuthorityRepository(t *testing.T) {
	ctxContainer, err := ops.ContainerDBStart("./../../..")
	assert.NoError(t, err)
	s := new(UserAuthorityRepositorySuite)
	suite.Run(t, s)
	ops.ContainerDBStop(ctxContainer)
}

func (s *UserAuthorityRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.db = database.NewPGDatabase()
	s.repo = NewUserAuthorityRepository(s.db)

	s.userAdminId, _ = uuid.Parse("8c61f220-8bb8-48b9-b225-d54dfa6503db")
}

func (s *UserAuthorityRepositorySuite) TearDownTest() {
	s.db.Close()
	s.mockCtrl.Finish()
}

func (s UserAuthorityRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	err := ops.DatasetTest(s.db, "./../../..", cleanScript, scripts...)
	s.NoError(err)
}

func (s *UserAuthorityRepositorySuite) TestFindAuthoritiesByUserIDUserExists() {
	s.prepareDatabase(true, "add-users.sql")
	a, err := s.repo.FindAuthoritiesByUserID(context.Background(), s.userAdminId)
	s.NoError(err)
	s.NotNil(a)
	s.Len(a, 2)
}

func (s *UserAuthorityRepositorySuite) TestFindAuthoritiesByUserIDUserNotExists() {
	s.prepareDatabase(true)
	a, err := s.repo.FindAuthoritiesByUserID(context.Background(), s.userAdminId)
	s.NoError(err)
	s.Nil(a)
}
