package repository

import (
	"github.com/stretchr/testify/assert"
	"golauth/config/datasource"
	"golauth/model"
	"testing"
)

func TestUserRoleRepository(t *testing.T) {
	ctx := Up()
	defer Down(ctx)

	dbTest, err := datasource.CreateDBConnection()
	if err != nil {
		t.Fatalf("error when creating datasource: %s", err.Error())
	}

	repo := NewUserRoleRepository(dbTest)

	t.Run("AddUserRole ok", func(t *testing.T) {
		u := model.User{
			Username:  "guest",
			FirstName: "Guest",
			LastName:  "None",
			Email:     "guest@none.com",
			Document:  "123456",
			Password:  "e10adc3949ba59abbe56e057f20f883e",
			Enabled:   true,
		}
		user, err := NewUserRepository(dbTest).Create(u)
		assert.NoError(t, err)
		assert.NotNil(t, user)

		role, err := NewRoleRepository(dbTest).FindByName("USER")
		assert.NoError(t, err)
		assert.NotNil(t, role)

		userRole, err := repo.AddUserRole(user.ID, role.ID)
		assert.NoError(t, err)
		assert.NotNil(t, userRole)
		assert.NotNil(t, userRole.CreationDate)
	})
}
