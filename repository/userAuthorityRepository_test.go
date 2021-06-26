package repository

import (
	"github.com/stretchr/testify/assert"
	"golauth/config/datasource"
	"testing"
)

func TestUserAuthorityRepository_FindAuthoritiesByUserID(t *testing.T) {
	ctx := Up()
	defer Down(ctx)

	dbTest, err := datasource.CreateDBConnection()
	if err != nil {
		t.Fatal("error when creating datasource: %w", err)
	}

	repo := NewUserAuthorityRepository(dbTest)

	t.Run("load authorities user exists", func(t *testing.T) {
		a, err := repo.FindAuthoritiesByUserID(1)
		assert.NoError(t, err)
		assert.NotNil(t, a)
		assert.Len(t, a, 2)
	})

	t.Run("load authorities user not exists", func(t *testing.T) {
		a, err := repo.FindAuthoritiesByUserID(999)
		assert.NoError(t, err)
		assert.Nil(t, a)
	})
}
