package repository

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golauth/config/datasource"
	"testing"
)

func TestUserAuthorityRepository(t *testing.T) {
	ctx := Up(true)
	defer Down(ctx)

	dbTest, err := datasource.CreateDBConnection()
	if err != nil {
		t.Fatal("error when creating datasource: %w", err)
	}

	repo := NewUserAuthorityRepository(dbTest)

	t.Run("FindAuthoritiesByUserID load authorities user exists", func(t *testing.T) {
		a, err := repo.FindAuthoritiesByUserID(1)
		assert.NoError(t, err)
		assert.NotNil(t, a)
		assert.Len(t, a, 2)
	})

	t.Run("FindAuthoritiesByUserID load authorities user not exists", func(t *testing.T) {
		a, err := repo.FindAuthoritiesByUserID(999)
		assert.NoError(t, err)
		assert.Nil(t, a)
	})
}

func TestUserAuthorityRepositoryWithMock(t *testing.T) {
	dbTest, mock := newDBMock()
	repo := NewUserAuthorityRepository(dbTest)
	defer func() {
		_ = dbTest.Close()
	}()

	t.Run("FindAuthoritiesByUserID with error not find by user", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(1).WillReturnError(mockDBClosedError)
		result, err := repo.FindAuthoritiesByUserID(1)
		assert.Empty(t, result)
		assert.Error(t, err)
		assert.ErrorAs(t, err, &mockDBClosedError)
	})

	t.Run("FindAuthoritiesByUserID error when parsing result to slice", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"name"}).
			AddRow("user").
			AddRow(nil).RowError(2, mockScanError)

		mock.ExpectQuery("SELECT").WithArgs(1).WillReturnRows(rows)
		result, err := repo.FindAuthoritiesByUserID(1)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.ErrorAs(t, err, &mockScanError)
	})
}
