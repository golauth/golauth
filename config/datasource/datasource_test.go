package datasource

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golauth/repository"
	"testing"
)

func TestDatabase_ok(t *testing.T) {
	ctx := repository.Up()

	testedDb, err := CreateDBConnection()
	if err != nil {
		t.Fatalf("error when stating database configuration: %s", err.Error())
	}
	assert.NotNil(t, testedDb, "database object not initialized")

	t.Run("table golauth_user exists", func(t *testing.T) {
		//var result int
		//err := testedDb.QueryRow("select 1 from information_schema.tables where table_schema = 'golauth' and table_name = 'golauth_user' ").
		//	Scan(&result)
		result, err := findTable(testedDb, "golauth_user")
		assert.NoError(t, err)
		assert.Equal(t, 1, result, "table users not created")
	})

	t.Run("table golauth_role exists", func(t *testing.T) {
		//var result int
		//err := testedDb.QueryRow("select 1 from information_schema.tables where table_schema = 'golauth' and table_name = 'golauth_role' ").
		//	Scan(&result)
		result, err := findTable(testedDb, "golauth_role")
		assert.NoError(t, err)
		assert.Equal(t, 1, result, "table users not created")
	})

	repository.Down(ctx)
}

func findTable(db *sql.DB, table string) (result int, err error) {
	err = db.QueryRow(fmt.Sprintf("select 1 from information_schema.tables where table_schema = 'golauth' and table_name = '%s'", table)).
		Scan(&result)
	return
}

func TestDatabase_nok(t *testing.T) {
	testedDb, err := CreateDBConnection()
	if err == nil || testedDb != nil {
		t.Fatal("database initialized without db instance started.")
	}
}
