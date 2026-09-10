package postgres

import (
	"context"
	"github.com/golauth/golauth/pkg/domain/apperr"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/golauth/golauth/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
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

// AdminExists drives the start-up bootstrap: true once a user holds the ADMIN
// authority (add-users seeds admin with the ADMIN role), false on an empty
// database.
func (s *UserRepositorySuite) TestAdminExists() {
	s.prepareDatabase(true, "add-users.sql")
	has, err := s.repo.AdminExists(context.Background())
	s.NoError(err)
	s.True(has)

	s.prepareDatabase(true) // clear-data only: no roles, no users
	has, err = s.repo.AdminExists(context.Background())
	s.NoError(err)
	s.False(has)
}

// A disabled admin membership does not count: the service is still without a
// usable administrator, so the bootstrap must run.
func (s *UserRepositorySuite) TestAdminExistsIgnoresDisabledMembership() {
	s.prepareDatabase(true, "add-users.sql")
	ctx := context.Background()

	_, err := s.db.Exec(ctx,
		`UPDATE golauth_user_role SET enabled = false
		 WHERE role_id = (SELECT id FROM golauth_role WHERE name = 'ADMIN')`)
	s.NoError(err)

	has, err := s.repo.AdminExists(ctx)
	s.NoError(err)
	s.False(has)
}

// A missing row is translated to apperr.ErrNotFound (mapped to HTTP 404), not a
// raw sql.ErrNoRows.
func (s *UserRepositorySuite) TestFindByIDMissingIsAppErrNotFound() {
	s.prepareDatabase(true, "add-users.sql")
	_, err := s.repo.FindByID(context.Background(), uuid.New())
	s.ErrorIs(err, apperr.ErrNotFound)
}

func (s *UserRepositorySuite) TestFindByUsernameMissingIsAppErrNotFound() {
	s.prepareDatabase(true, "add-users.sql")
	_, err := s.repo.FindByUsername(context.Background(), "nobody")
	s.ErrorIs(err, apperr.ErrNotFound)
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

// A second insert colliding with ui_golauth_user_username surfaces as the
// domain ErrUserAlreadyExists (mapped to HTTP 409), not a raw driver error.
func (s *UserRepositorySuite) TestCreateDuplicateUsernameIsAlreadyExists() {
	s.prepareDatabase(true, "add-users.sql")
	u := &entity.User{
		Username: "admin", FirstName: "Dup", LastName: "User",
		Email: "different@none.com", Document: "999", Password: "irrelevant",
	}

	_, err := s.repo.Create(context.Background(), u)
	s.ErrorIs(err, repository.ErrUserAlreadyExists)
}

// Same guarantee for the ui_golauth_user_email unique index.
func (s *UserRepositorySuite) TestCreateDuplicateEmailIsAlreadyExists() {
	s.prepareDatabase(true, "add-users.sql")
	u := &entity.User{
		Username: "brand-new", FirstName: "Dup", LastName: "User",
		Email: "admin@goauth.org", Document: "999", Password: "irrelevant",
	}

	_, err := s.repo.Create(context.Background(), u)
	s.ErrorIs(err, repository.ErrUserAlreadyExists)
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
