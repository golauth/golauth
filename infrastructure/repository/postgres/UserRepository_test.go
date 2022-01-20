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
	"time"
)

type UserRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       *sql.DB

	repo repository.UserRepository
}

func TestUserRepository(t *testing.T) {
	ctxContainer, err := ops.ContainerDBStart("./../../..")
	assert.NoError(t, err)
	s := new(UserRepositorySuite)
	suite.Run(t, s)
	ops.ContainerDBStop(ctxContainer)
}

func (s *UserRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	ds, err := datasource.NewDatasource()
	s.NotNil(ds)
	s.NoError(err)
	s.db = ds.GetDB()

	s.repo = NewUserRepository(s.db)
}

func (s *UserRepositorySuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s UserRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	err := ops.DatasetTest(s.db, "./../../..", cleanScript, scripts...)
	s.NoError(err)
}

func (s *UserRepositorySuite) TestFindUserWithoutPassword() {
	s.prepareDatabase(true, "add-users.sql")
	id, _ := uuid.Parse("8c61f220-8bb8-48b9-b225-d54dfa6503db")
	u, err := s.repo.FindByID(id)
	s.NoError(err)
	s.NotNil(u)
	s.Equal("admin", u.Username)
	s.Empty(u.Password)
}

func (s *UserRepositorySuite) TestFindUserWithPassword() {
	s.prepareDatabase(true, "add-users.sql")
	u, err := s.repo.FindByUsername("admin")
	s.NoError(err)
	s.NotNil(u)
	s.Equal("admin", u.Username)
	s.NotEmpty(u.Password)
}

func (s *UserRepositorySuite) TestFindUserByIdWithoutPassword() {
	s.prepareDatabase(true, "add-users.sql")
	userId, _ := uuid.Parse("8c61f220-8bb8-48b9-b225-d54dfa6503db")
	u, err := s.repo.FindByID(userId)
	s.NoError(err)
	s.NotNil(u)
	s.Equal("admin", u.Username)
	s.Zero(u.Password)
}

func (s *UserRepositorySuite) TestCreateNewUserOk() {
	s.prepareDatabase(true, "add-users.sql")
	u := entity.User{
		Username:     "guest",
		FirstName:    "Guest",
		LastName:     "None",
		Email:        "guest@none.com",
		Document:     "123456",
		Password:     "e10adc3949ba59abbe56e057f20f883e",
		Enabled:      true,
		CreationDate: time.Now(),
	}

	user, err := s.repo.Create(u)
	s.NoError(err)
	s.NotEmpty(user.ID)
}

// =====================================================================================
type UserRepositoryDBMockSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       *sql.DB
	mockDB   sqlmock.Sqlmock
	repo     repository.UserRepository
}

func TestUserRepositoryWithMock(t *testing.T) {
	suite.Run(t, new(UserRepositoryDBMockSuite))
}

func (s *UserRepositoryDBMockSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	var err error
	s.db, s.mockDB, err = sqlmock.New()
	s.NoError(err)

	s.repo = NewUserRepository(s.db)
}

func (s *UserRepositoryDBMockSuite) TearDownTest() {
	_ = s.db.Close()
	s.mockCtrl.Finish()
}

func (s *UserRepositoryDBMockSuite) TestFindByUsernameScanError() {
	s.mockDB.ExpectQuery("SELECT").
		WithArgs("username").
		WillReturnError(ops.ErrMockScan)
	result, err := s.repo.FindByUsername("username")
	s.Empty(result)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockScan)
}

func (s *UserRepositoryDBMockSuite) TestFindByIDScanError() {
	s.mockDB.ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(ops.ErrMockScan)
	result, err := s.repo.FindByID(uuid.New())
	s.Empty(result)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockScan)
}

func (s *UserRepositoryDBMockSuite) TestCreateScanError() {
	s.mockDB.ExpectQuery("INSERT").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(ops.ErrMockScan)
	result, err := s.repo.Create(entity.User{Username: "username"})
	s.Empty(result)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockScan)
}
