package postgres

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/domain/entity"
	"golauth/domain/repository"
	"golauth/infrastructure/datasource"
	"golauth/ops"
	"testing"
)

type UserAuthorityRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       *sql.DB

	repo repository.UserAuthorityRepository

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
	ds, err := datasource.NewDatasource()
	s.NotNil(ds)
	s.NoError(err)
	s.db = ds.GetDB()

	s.repo = NewUserAuthorityRepository(s.db)

	s.userAdminId, _ = uuid.Parse("8c61f220-8bb8-48b9-b225-d54dfa6503db")
}

func (s *UserAuthorityRepositorySuite) TearDownTest() {
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
	a, err := s.repo.FindAuthoritiesByUserID(s.userAdminId)
	s.NoError(err)
	s.NotNil(a)
	s.Len(a, 2)
}

func (s *UserAuthorityRepositorySuite) TestFindAuthoritiesByUserIDUserNotExists() {
	s.prepareDatabase(true)
	a, err := s.repo.FindAuthoritiesByUserID(s.userAdminId)
	s.NoError(err)
	s.Nil(a)
}

// =====================================================================================
type UserAuthorityRepositoryDBMockSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl    *gomock.Controller
	db          *sql.DB
	mockDB      sqlmock.Sqlmock
	repo        repository.UserAuthorityRepository
	roleMock    entity.Role
	userAdminId uuid.UUID
}

func TestUserAuthorityRepositoryWithMock(t *testing.T) {
	suite.Run(t, new(UserAuthorityRepositoryDBMockSuite))
}

func (s *UserAuthorityRepositoryDBMockSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	var err error
	s.db, s.mockDB, err = sqlmock.New()
	s.NoError(err)

	s.repo = NewUserAuthorityRepository(s.db)

	s.roleMock = entity.Role{Name: "role", Description: "role", Enabled: true}
	s.userAdminId, _ = uuid.Parse("8c61f220-8bb8-48b9-b225-d54dfa6503db")
}

func (s *UserAuthorityRepositoryDBMockSuite) TearDownTest() {
	_ = s.db.Close()
	s.mockCtrl.Finish()
}

func (s *UserAuthorityRepositoryDBMockSuite) TestUserAuthorityRepositoryWithMockErrDbClosed() {
	s.mockDB.ExpectQuery("SELECT").WithArgs(s.userAdminId).WillReturnError(ops.ErrMockDBClosed)
	result, err := s.repo.FindAuthoritiesByUserID(s.userAdminId)
	s.Empty(result)
	s.Error(err)
	s.ErrorAs(err, &ops.ErrMockDBClosed)
}

func (s *UserAuthorityRepositoryDBMockSuite) TestUserAuthorityRepositoryWithMockScanErr() {
	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("user").
		AddRow(nil).RowError(2, ops.ErrMockScan)

	s.mockDB.ExpectQuery("SELECT").WithArgs(s.userAdminId).WillReturnRows(rows)
	result, err := s.repo.FindAuthoritiesByUserID(s.userAdminId)
	s.Error(err)
	s.Empty(result)
	s.ErrorAs(err, &ops.ErrMockScan)
}
