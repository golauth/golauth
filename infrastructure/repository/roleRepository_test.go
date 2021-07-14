package repository

import (
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/entity"
	datasource2 "golauth/infrastructure/datasource"
	"golauth/ops"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type RoleRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       *sql.DB

	repo RoleRepository
}

func TestRoleRepository(t *testing.T) {
	ctxContainer, err := ops.ContainerDBStart("./../..")
	assert.NoError(t, err)
	s := new(RoleRepositorySuite)
	suite.Run(t, s)
	ops.ContainerDBStop(ctxContainer)
}

func (s *RoleRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	ds, err := datasource2.NewDatasource()
	s.NotNil(ds)
	s.NoError(err)
	s.db = ds.GetDB()

	s.repo = NewRoleRepository(s.db)
}

func (s *RoleRepositorySuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s RoleRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	err := ops.DatasetTest(s.db, "./../..", cleanScript, scripts...)
	s.NoError(err)
}

func (s RoleRepositorySuite) TestRoleRepositoryFindRoleByName() {
	s.prepareDatabase(true, "add-users.sql")
	role, err := s.repo.FindByName("USER")
	s.NoError(err)
	s.NotNil(role)
	s.Equal("USER", role.Name)
}

func (s RoleRepositorySuite) TestRoleRepositoryCreateNewRole() {
	s.prepareDatabase(true, "add-users.sql")
	r := entity.Role{
		Name:         "CUSTOMER_EDIT",
		Description:  "Customer edit",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	role, err := s.repo.Create(r)
	s.NoError(err)
	s.NotNil(role)
	s.NotNil(role.ID)
	s.Equal("CUSTOMER_EDIT", role.Name)
}

func (s RoleRepositorySuite) TestRoleRepositoryEditOk() {
	s.prepareDatabase(true, "add-users.sql")
	r, err := s.repo.FindByName("USER")
	s.NoError(err)
	s.NotNil(r)
	s.Equal("Role USER", r.Description)

	r.Description = "Role to common user"
	err = s.repo.Edit(r)
	s.NoError(err)

	edited, err := s.repo.FindByName("USER")
	s.NoError(err)
	s.NotNil(edited)
	s.Equal("Role to common user", edited.Description)
}

func (s RoleRepositorySuite) TestRoleRepositoryEditIdNotFound() {
	s.prepareDatabase(true)
	r := entity.Role{
		ID:           uuid.New(),
		Name:         "CUSTOMER_EDIT",
		Description:  "Customer edit",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	err := s.repo.Edit(r)
	s.Error(err)
	expectedErr := errors.New("no rows affected")
	s.ErrorAs(err, &expectedErr)
}

func (s RoleRepositorySuite) TestRoleRepositoryFindByNameNotFound() {
	s.prepareDatabase(true)
	role, err := s.repo.FindByName("USER")
	s.Empty(role)
	s.NotNil(err)
	expectedErr := errors.New("could not find role USER")
	s.ErrorAs(err, &expectedErr)
}

func (s RoleRepositorySuite) TestRoleRepositoryCreateDuplicatedRole() {
	s.prepareDatabase(true, "add-users.sql")
	r := entity.Role{
		Name:         "USER",
		Description:  "New Role User",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	role, err := s.repo.Create(r)
	s.Empty(role)
	s.NotNil(err)
	expectedErr := errors.New("could not create role USER")
	s.ErrorAs(err, &expectedErr)
}

func (s RoleRepositorySuite) TestRoleRepositoryChangeStatusOk() {
	s.prepareDatabase(true, "add-users.sql")
	id, _ := uuid.Parse("c12b415b-c3ad-487f-9800-f548aa18cc58")
	err := s.repo.ChangeStatus(id, false)
	s.NoError(err)

	edited, err := s.repo.FindByName("USER")
	s.NoError(err)
	s.NotNil(edited)
	s.False(edited.Enabled)
}

func (s RoleRepositorySuite) TestRoleRepositoryExistsByID() {
	s.prepareDatabase(true, "add-users.sql")
	id, _ := uuid.Parse("c12b415b-c3ad-487f-9800-f548aa18cc58")
	exists, err := s.repo.ExistsById(id)
	s.NoError(err)
	s.True(exists)
}

func (s RoleRepositorySuite) TestRoleRepositoryNotExistsByID() {
	s.prepareDatabase(true, "add-users.sql")
	exists, err := s.repo.ExistsById(uuid.New())
	s.NoError(err)
	s.False(exists)
}

// =====================================================================================

type RoleRepositoryDBMockSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       *sql.DB
	mockDB   sqlmock.Sqlmock
	repo     RoleRepository
	roleMock entity.Role
}

func TestRoleRepositoryWithMock(t *testing.T) {
	suite.Run(t, new(RoleRepositoryDBMockSuite))
}

func (s *RoleRepositoryDBMockSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	var err error
	s.db, s.mockDB, err = sqlmock.New()
	s.NoError(err)

	s.repo = NewRoleRepository(s.db)

	s.roleMock = entity.Role{Name: "role", Description: "role", Enabled: true}
}

func (s *RoleRepositoryDBMockSuite) TearDownTest() {
	_ = s.db.Close()
	s.mockCtrl.Finish()
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockFindScanError() {
	s.mockDB.ExpectQuery("SELECT").
		WithArgs("role").
		WillReturnError(ops.ErrMockScan)
	result, err := s.repo.FindByName("role")
	s.Empty(result)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockScan)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockCreateScanError() {
	s.mockDB.ExpectQuery("INSERT").
		WithArgs(s.roleMock.Name, s.roleMock.Description, s.roleMock.Enabled).
		WillReturnError(ops.ErrMockScan)
	result, err := s.repo.Create(s.roleMock)
	s.Empty(result)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockScan)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockEditExecError() {
	s.mockDB.ExpectExec("UPDATE").
		WithArgs(s.roleMock.ID, s.roleMock.Name, s.roleMock.Description, s.roleMock.Enabled).
		WillReturnError(ops.ErrMockUpdate)
	err := s.repo.Edit(s.roleMock)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockUpdate)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockEditNoRowsAffected() {
	s.mockDB.ExpectExec("UPDATE").
		WithArgs(s.roleMock.ID, s.roleMock.Name, s.roleMock.Description, s.roleMock.Enabled).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := s.repo.Edit(s.roleMock)
	s.NotNil(err)
	s.ErrorAs(err, &sql.ErrNoRows)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockChangeStatusExecError() {
	id := uuid.New()
	s.mockDB.ExpectExec("UPDATE").
		WithArgs(id, false).
		WillReturnError(ops.ErrMockUpdate)
	err := s.repo.ChangeStatus(id, false)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockUpdate)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockChangeNoRowsAffected() {
	id := uuid.New()
	s.mockDB.ExpectExec("UPDATE").
		WithArgs(id, false).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := s.repo.ChangeStatus(id, false)
	s.NotNil(err)
	s.ErrorAs(err, &sql.ErrNoRows)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockExistsByIDScanError() {
	id := uuid.New()
	s.mockDB.ExpectQuery("SELECT EXISTS").
		WithArgs(id).
		WillReturnError(ops.ErrMockScan)
	result, err := s.repo.ExistsById(id)
	s.Empty(result)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockScan)
}
