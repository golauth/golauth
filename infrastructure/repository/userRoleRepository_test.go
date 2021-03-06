package repository

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/entity"
	"golauth/infrastructure/datasource"
	"golauth/ops"
	"testing"
)

type UserRoleRepositorySuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       *sql.DB

	repo UserRoleRepository
}

func TestUserRoleRepository(t *testing.T) {
	ctxContainer, err := ops.ContainerDBStart("./../..")
	assert.NoError(t, err)
	s := new(UserRoleRepositorySuite)
	suite.Run(t, s)
	ops.ContainerDBStop(ctxContainer)
}

func (s *UserRoleRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	ds, err := datasource.NewDatasource()
	s.NotNil(ds)
	s.NoError(err)
	s.db = ds.GetDB()

	s.repo = NewUserRoleRepository(s.db)
}

func (s *UserRoleRepositorySuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s UserRoleRepositorySuite) prepareDatabase(clean bool, scripts ...string) {
	cleanScript := ""
	if clean {
		cleanScript = "clear-data.sql"
	}
	err := ops.DatasetTest(s.db, "./../..", cleanScript, scripts...)
	s.NoError(err)
}

func (s *UserRoleRepositorySuite) TestAddUserRole() {
	u := entity.User{
		Username:  "guest",
		FirstName: "Guest",
		LastName:  "None",
		Email:     "guest@none.com",
		Document:  "123456",
		Password:  "e10adc3949ba59abbe56e057f20f883e",
		Enabled:   true,
	}
	user, err := NewUserRepository(s.db).Create(u)
	s.NoError(err)
	s.NotNil(user)

	role, err := NewRoleRepository(s.db).FindByName("USER")
	s.NoError(err)
	s.NotNil(role)

	err = s.repo.AddUserRole(user.ID, role.ID)
	s.NoError(err)
}

// =====================================================================================

type UserRoleRepositoryDBMockSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl     *gomock.Controller
	db           *sql.DB
	mockDB       sqlmock.Sqlmock
	repo         UserRoleRepository
	roleMock     entity.Role
	userAdmin2Id uuid.UUID
	roleAdminId  uuid.UUID
}

func TestUserRoleRepositoryWithMock(t *testing.T) {
	suite.Run(t, new(UserRoleRepositoryDBMockSuite))
}

func (s *UserRoleRepositoryDBMockSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	var err error
	s.db, s.mockDB, err = sqlmock.New()
	s.NoError(err)

	s.repo = NewUserRoleRepository(s.db)
	s.userAdmin2Id, _ = uuid.Parse("e227d878-b5d6-4902-a500-3357955c962d")
	s.roleAdminId, _ = uuid.Parse("7f68301e-df80-45bd-9532-23a58733ef2c")
}

func (s *UserRoleRepositoryDBMockSuite) TearDownTest() {
	_ = s.db.Close()
	s.mockCtrl.Finish()
}

func (s UserRoleRepositoryDBMockSuite) TestRoleRepositoryWithMockAddUserRoleScanError() {
	s.mockDB.ExpectQuery("INSERT INTO golauth_user_role").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(ops.ErrMockScan)
	err := s.repo.AddUserRole(s.userAdmin2Id, s.roleAdminId)
	s.NotNil(err)
	s.ErrorAs(err, &ops.ErrMockScan)
}
