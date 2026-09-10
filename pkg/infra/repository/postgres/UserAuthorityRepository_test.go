package postgres

import (
	"context"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/golauth/golauth/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
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
	ctxContainer, err := tests.ContainerDBStart("./../../../..")
	assert.NoError(t, err)
	s := new(UserAuthorityRepositorySuite)
	suite.Run(t, s)
	tests.ContainerDBStop(ctxContainer)
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

func (s *UserAuthorityRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	err := tests.DatasetTest(s.db, "./../../../..", cleanScript, scripts...)
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

// TestFindAuthoritiesExcludesDisabledRoleAndAuthority is the observable effect
// of PATCH /auth/roles/:id/change-status: a disabled role grants nothing, and
// re-enabling it (through the same ChangeStatus the endpoint calls) brings the
// authority back. A disabled authority is filtered too, even under an enabled
// role.
func (s *UserAuthorityRepositorySuite) TestFindAuthoritiesExcludesDisabledRoleAndAuthority() {
	s.prepareDatabase(true, "add-user-disabled-role.sql")
	ctx := context.Background()
	adminRoleID, _ := uuid.Parse("7f68301e-df80-45bd-9532-23a58733ef2c")
	roleRepo := NewRoleRepository(s.db)

	a, err := s.repo.FindAuthoritiesByUserID(ctx, s.userAdminId)
	s.NoError(err)
	s.Empty(a)

	s.NoError(roleRepo.ChangeStatus(ctx, adminRoleID, true))
	a, err = s.repo.FindAuthoritiesByUserID(ctx, s.userAdminId)
	s.NoError(err)
	s.Equal([]string{"ADMIN"}, a)

	_, err = s.db.Exec(ctx, "UPDATE golauth_authority SET enabled = false WHERE name = 'ADMIN'")
	s.NoError(err)
	a, err = s.repo.FindAuthoritiesByUserID(ctx, s.userAdminId)
	s.NoError(err)
	s.Empty(a)
}

// A disabled row in golauth_user_role grants nothing, even when the role and
// the authority are both enabled. This is the enabled column added in the
// referential-integrity migration feeding the same WHERE clause.
func (s *UserAuthorityRepositorySuite) TestFindAuthoritiesExcludesDisabledMembership() {
	s.prepareDatabase(true, "add-users.sql")
	ctx := context.Background()

	a, err := s.repo.FindAuthoritiesByUserID(ctx, s.userAdminId)
	s.NoError(err)
	s.ElementsMatch([]string{"ADMIN", "USER"}, a)

	_, err = s.db.Exec(ctx,
		`UPDATE golauth_user_role SET enabled = false
		 WHERE user_id = $1 AND role_id = (SELECT id FROM golauth_role WHERE name = 'USER')`,
		s.userAdminId)
	s.NoError(err)

	a, err = s.repo.FindAuthoritiesByUserID(ctx, s.userAdminId)
	s.NoError(err)
	s.Equal([]string{"ADMIN"}, a)
}
