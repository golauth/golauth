package postgres

import (
	"context"
	"testing"

	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/golauth/golauth/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

const (
	seedAdminUserID = "8c61f220-8bb8-48b9-b225-d54dfa6503db"
	seedUserRoleID  = "c12b415b-c3ad-487f-9800-f548aa18cc58"
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

func (s *UserRoleRepositorySuite) prepareDatabase(scripts ...string) {
	s.NoError(tests.DatasetTest(s.db, "./../../../..", "clear-data.sql", scripts...))
}

// A membership pointing at a role that does not exist is rejected by the
// foreign key, at the database level -- the application never checked.
func (s *UserRoleRepositorySuite) TestAddUserRoleForMissingRoleFails() {
	s.prepareDatabase("add-users.sql")
	err := s.repo.AddUserRole(context.Background(), uuid.MustParse(seedAdminUserID), uuid.New())
	s.Error(err)
}

func (s *UserRoleRepositorySuite) TestAddUserRoleForMissingUserFails() {
	s.prepareDatabase("add-users.sql")
	err := s.repo.AddUserRole(context.Background(), uuid.New(), uuid.MustParse(seedUserRoleID))
	s.Error(err)
}

// Deleting a user cascades: its memberships go with it.
func (s *UserRoleRepositorySuite) TestDeletingUserCascadesMemberships() {
	s.prepareDatabase("add-users.sql")
	ctx := context.Background()

	var before int
	s.NoError(s.db.One(ctx, "SELECT count(*) FROM golauth_user_role WHERE user_id = $1", seedAdminUserID).Scan(&before))
	s.Positive(before)

	_, err := s.db.Exec(ctx, "DELETE FROM golauth_user WHERE id = $1", seedAdminUserID)
	s.NoError(err)

	var after int
	s.NoError(s.db.One(ctx, "SELECT count(*) FROM golauth_user_role WHERE user_id = $1", seedAdminUserID).Scan(&after))
	s.Zero(after)
}

// Deleting a role that still has members is refused (ON DELETE RESTRICT): it
// must be a deliberate act.
func (s *UserRoleRepositorySuite) TestDeletingRoleInUseIsRefused() {
	s.prepareDatabase("add-users.sql")
	_, err := s.db.Exec(context.Background(), "DELETE FROM golauth_role WHERE id = $1", seedUserRoleID)
	s.Error(err)
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
