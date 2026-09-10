package postgres

import (
	"context"
	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/domain/entity"
	"github.com/golauth/golauth/internal/domain/repository"
	"github.com/golauth/golauth/internal/infra/database"
	"github.com/golauth/golauth/internal/testsupport"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type RoleRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       database.Database

	repo repository.RoleRepository
}

func TestRoleRepository(t *testing.T) {
	ctxContainer, err := testsupport.ContainerDBStart("./../../../..")
	assert.NoError(t, err)
	s := new(RoleRepositorySuite)
	suite.Run(t, s)
	testsupport.ContainerDBStop(ctxContainer)
}

func (s *RoleRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.db = database.NewPGDatabase()
	s.NotNil(s.db)
	s.repo = NewRoleRepository(s.db)
}

func (s *RoleRepositorySuite) TearDownTest() {
	s.db.Close()
	s.mockCtrl.Finish()
}

func (s *RoleRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	err := testsupport.DatasetTest(s.db, "./../../../..", cleanScript, scripts...)
	s.NoError(err)
}

func (s *RoleRepositorySuite) TestRoleRepositoryFindRoleByName() {
	s.prepareDatabase(true, "add-users.sql")
	role, err := s.repo.FindByName(context.Background(), "USER")
	s.NoError(err)
	s.NotNil(role)
	s.Equal("USER", role.Name)
}

func (s *RoleRepositorySuite) TestRoleRepositoryCreateNewRole() {
	s.prepareDatabase(true, "add-users.sql")
	r := &entity.Role{
		Name:         "CUSTOMER_EDIT",
		Description:  "Customer edit",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	saved, err := s.repo.Create(context.Background(), r)
	s.NoError(err)
	s.NotNil(saved)
	s.NotNil(saved.ID)
	s.Equal("CUSTOMER_EDIT", saved.Name)
}

func (s *RoleRepositorySuite) TestRoleRepositoryEditOk() {
	s.prepareDatabase(true, "add-users.sql")
	r, err := s.repo.FindByName(context.Background(), "USER")
	s.NoError(err)
	s.NotNil(r)
	s.Equal("Role USER", r.Description)

	r.Description = "Role to common user"
	err = s.repo.Edit(context.Background(), r)
	s.NoError(err)

	edited, err := s.repo.FindByName(context.Background(), "USER")
	s.NoError(err)
	s.NotNil(edited)
	s.Equal("Role to common user", edited.Description)
}

func (s *RoleRepositorySuite) TestRoleRepositoryEditIdNotFound() {
	s.prepareDatabase(true)
	r := &entity.Role{
		ID:           uuid.New(),
		Name:         "CUSTOMER_EDIT",
		Description:  "Customer edit",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	err := s.repo.Edit(context.Background(), r)
	s.ErrorIs(err, apperr.ErrNotFound)
}

func (s *RoleRepositorySuite) TestRoleRepositoryFindByNameNotFound() {
	s.prepareDatabase(true)
	role, err := s.repo.FindByName(context.Background(), "USER")
	s.Empty(role)
	s.ErrorIs(err, apperr.ErrNotFound)
}

func (s *RoleRepositorySuite) TestRoleRepositoryCreateDuplicatedRole() {
	s.prepareDatabase(true, "add-users.sql")
	r := &entity.Role{
		Name:         "USER",
		Description:  "New Role User",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	role, err := s.repo.Create(context.Background(), r)
	s.Empty(role)
	s.ErrorIs(err, apperr.ErrAlreadyExists)
}

func (s *RoleRepositorySuite) TestRoleRepositoryChangeStatusOk() {
	s.prepareDatabase(true, "add-users.sql")
	id, _ := uuid.Parse("c12b415b-c3ad-487f-9800-f548aa18cc58")
	err := s.repo.ChangeStatus(context.Background(), id, false)
	s.NoError(err)

	edited, err := s.repo.FindByName(context.Background(), "USER")
	s.NoError(err)
	s.NotNil(edited)
	s.False(edited.Enabled)
}

// A ChangeStatus against an id that matches no row is not-found (Plan 08 maps it
// to 404), not a silent success and not a wrapped nil error.
func (s *RoleRepositorySuite) TestRoleRepositoryChangeStatusIdNotFound() {
	s.prepareDatabase(true)
	err := s.repo.ChangeStatus(context.Background(), uuid.New(), false)
	s.ErrorIs(err, apperr.ErrNotFound)
}

func (s *RoleRepositorySuite) TestRoleRepositoryExistsByID() {
	s.prepareDatabase(true, "add-users.sql")
	id, _ := uuid.Parse("c12b415b-c3ad-487f-9800-f548aa18cc58")
	exists, err := s.repo.ExistsById(context.Background(), id)
	s.NoError(err)
	s.True(exists)
}

func (s *RoleRepositorySuite) TestRoleRepositoryNotExistsByID() {
	s.prepareDatabase(true, "add-users.sql")
	exists, err := s.repo.ExistsById(context.Background(), uuid.New())
	s.NoError(err)
	s.False(exists)
}
