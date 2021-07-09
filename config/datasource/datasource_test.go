package datasource

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/postgrescontainer"
	"testing"
)

type DatasourceSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	ds Datasource
}

func TestDatasource(t *testing.T) {
	ctxContainer, err := postgrescontainer.ContainerDBStart("./../..")
	assert.NoError(t, err)
	s := new(DatasourceSuite)
	suite.Run(t, s)
	postgrescontainer.ContainerDBStop(ctxContainer)
}

func (s *DatasourceSuite) SetupTest() {
	var err error
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ds, err = NewDatasource()
	s.NoError(err)
	s.NotNil(s.ds)
}

func (s *DatasourceSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *DatasourceSuite) TestTableUserCreated() {
	fmt.Println("table golauth_user exists")
	result, err := s.findTable(s.ds.GetDB(), "golauth_user")
	s.NoError(err)
	s.Equal(1, result, "table users not created")
}

func (s *DatasourceSuite) TestTableRoleCreated() {
	fmt.Println("table golauth_role exists")
	result, err := s.findTable(s.ds.GetDB(), "golauth_role")
	s.NoError(err)
	s.Equal(1, result, "table users not created")
}

func (s *DatasourceSuite) findTable(db *sql.DB, table string) (result int, err error) {
	fmt.Printf("findTable=%s\n", table)
	err = db.QueryRow(fmt.Sprintf("select 1 from information_schema.tables where table_schema = 'golauth' and table_name = '%s'", table)).
		Scan(&result)
	return
}

// =====================================================================================

func TestDatasourceWithoutMigrations(t *testing.T) {
	ctxContainer, err := postgrescontainer.ContainerDBStart("./..")
	assert.NoError(t, err)
	_, err = NewDatasource()
	assert.NoError(t, err)
	postgrescontainer.ContainerDBStop(ctxContainer)
}

func TestDatasourceMigrationsTwice(t *testing.T) {
	ctxContainer, err := postgrescontainer.ContainerDBStart("./../..")
	assert.NoError(t, err)

	_, err = NewDatasource()
	assert.NoError(t, err)
	_, err = NewDatasource()
	assert.NoError(t, err)

	postgrescontainer.ContainerDBStop(ctxContainer)
}

func TestDatasourceConnectionCouldNotEstablished(t *testing.T) {
	ds, err := NewDatasource()
	assert.Nil(t, ds)
	assert.NotNil(t, err)
	expectedErr := errors.New("could not establish connection")
	assert.ErrorAs(t, err, &expectedErr)
}
