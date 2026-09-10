package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/golauth/golauth/internal/domain/repository"
	"github.com/golauth/golauth/internal/infra/database"
	"github.com/golauth/golauth/internal/testsupport"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type LoginAttemptRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       database.Database
	repo     repository.LoginAttemptRepository

	userAdminId uuid.UUID
}

func TestLoginAttemptRepository(t *testing.T) {
	ctxContainer, err := testsupport.ContainerDBStart("./../../../..")
	assert.NoError(t, err)
	s := new(LoginAttemptRepositorySuite)
	suite.Run(t, s)
	testsupport.ContainerDBStop(ctxContainer)
}

func (s *LoginAttemptRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.db = database.NewPGDatabase()
	s.repo = NewLoginAttemptRepository(s.db)
	s.userAdminId, _ = uuid.Parse("8c61f220-8bb8-48b9-b225-d54dfa6503db")
}

func (s *LoginAttemptRepositorySuite) TearDownTest() {
	s.db.Close()
	s.mockCtrl.Finish()
}

func (s *LoginAttemptRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	s.NoError(testsupport.DatasetTest(s.db, "./../../../..", cleanScript, scripts...))
}

// TestFailureLifecycle walks the whole counter lifecycle the token use case
// drives: no row, first failure, a second failure that also locks, and a reset
// that clears everything.
func (s *LoginAttemptRepositorySuite) TestFailureLifecycle() {
	ctx := context.Background()
	s.prepareDatabase(true, "add-users.sql")

	got, err := s.repo.Get(ctx, s.userAdminId)
	s.NoError(err)
	s.Nil(got, "a user with no failures has no row")

	s.NoError(s.repo.RegisterFailure(ctx, s.userAdminId, time.Time{}))
	got, err = s.repo.Get(ctx, s.userAdminId)
	s.NoError(err)
	s.Require().NotNil(got)
	s.Equal(1, got.FailedCount)
	s.False(got.Locked(time.Now()))

	s.NoError(s.repo.RegisterFailure(ctx, s.userAdminId, time.Now().Add(time.Hour)))
	got, err = s.repo.Get(ctx, s.userAdminId)
	s.NoError(err)
	s.Require().NotNil(got)
	s.Equal(2, got.FailedCount, "the second failure increments the same row")
	s.True(got.Locked(time.Now()), "and applies the lock window")

	s.NoError(s.repo.Reset(ctx, s.userAdminId))
	got, err = s.repo.Get(ctx, s.userAdminId)
	s.NoError(err)
	s.Nil(got, "reset removes the row")
}

// TestResetWithoutRowIsNoError: a successful first-ever login calls Reset on a
// user that has no row; that must be harmless.
func (s *LoginAttemptRepositorySuite) TestResetWithoutRowIsNoError() {
	s.prepareDatabase(true, "add-users.sql")
	s.NoError(s.repo.Reset(context.Background(), s.userAdminId))
}
