package repository

import (
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/config/datasource"
	"golauth/model"
	"golauth/postgrescontainer"
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
	ctxContainer, err := postgrescontainer.ContainerDBStart("./..")
	assert.NoError(t, err)
	s := new(RoleRepositorySuite)
	suite.Run(t, s)
	postgrescontainer.ContainerDBStop(ctxContainer)
}

func (s *RoleRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	ds, err := datasource.NewDatasource()
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
	err := postgrescontainer.DatasetTest(s.db, "./..", cleanScript, scripts...)
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
	r := model.Role{
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
	r := model.Role{
		ID:           1,
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
	r := model.Role{
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

// =====================================================================================

type RoleRepositoryDBMockSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       *sql.DB
	mockDB   sqlmock.Sqlmock
	repo     RoleRepository
	roleMock model.Role
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

	s.roleMock = model.Role{Name: "role", Description: "role", Enabled: true}
}

func (s *RoleRepositoryDBMockSuite) TearDownTest() {
	_ = s.db.Close()
	s.mockCtrl.Finish()
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockFindScanError() {
	s.mockDB.ExpectQuery("SELECT").
		WithArgs("role").
		WillReturnError(postgrescontainer.ErrMockScan)
	result, err := s.repo.FindByName("role")
	s.Empty(result)
	s.NotNil(err)
	s.ErrorAs(err, &postgrescontainer.ErrMockScan)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockCreateScanError() {
	s.mockDB.ExpectQuery("INSERT").
		WithArgs(s.roleMock.Name, s.roleMock.Description, s.roleMock.Enabled).
		WillReturnError(postgrescontainer.ErrMockScan)
	result, err := s.repo.Create(s.roleMock)
	s.Empty(result)
	s.NotNil(err)
	s.ErrorAs(err, &postgrescontainer.ErrMockScan)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockEditExecError() {
	s.mockDB.ExpectExec("UPDATE").
		WithArgs(s.roleMock.ID, s.roleMock.Name, s.roleMock.Description, s.roleMock.Enabled).
		WillReturnError(postgrescontainer.ErrMockUpdate)
	err := s.repo.Edit(s.roleMock)
	s.NotNil(err)
	s.ErrorAs(err, &postgrescontainer.ErrMockUpdate)
}

func (s RoleRepositoryDBMockSuite) TestRoleRepositoryWithMockEditNoRowsAffected() {
	s.mockDB.ExpectExec("UPDATE").
		WithArgs(s.roleMock.ID, s.roleMock.Name, s.roleMock.Description, s.roleMock.Enabled).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := s.repo.Edit(s.roleMock)
	s.NotNil(err)
	s.ErrorAs(err, &sql.ErrNoRows)
}
