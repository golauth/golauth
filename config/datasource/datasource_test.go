package datasource

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golauth/repository"
	"testing"
)

func TestDatasource_ok(t *testing.T) {
	ctx := repository.Up(true)

	ds, err := NewDatasource()
	assert.Nil(t, err)
	testedDb := ds.GetDB()
	assert.NotNil(t, testedDb, "database object not initialized")

	t.Run("table golauth_user exists", func(t *testing.T) {
		fmt.Println("table golauth_user exists")
		result, err := findTable(testedDb, "golauth_user")
		assert.NoError(t, err)
		assert.Equal(t, 1, result, "table users not created")
	})

	t.Run("table golauth_role exists", func(t *testing.T) {
		fmt.Println("table golauth_role exists")
		result, err := findTable(testedDb, "golauth_role")
		assert.NoError(t, err)
		assert.Equal(t, 1, result, "table users not created")
	})

	repository.Down(ctx)
}

func findTable(db *sql.DB, table string) (result int, err error) {
	fmt.Printf("findTable=%s\n", table)
	err = db.QueryRow(fmt.Sprintf("select 1 from information_schema.tables where table_schema = 'golauth' and table_name = '%s'", table)).
		Scan(&result)
	return
}

func TestDatasource_nok(t *testing.T) {
	ds, err := NewDatasource()
	assert.Nil(t, ds)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "could not establish connection:")
}
