package repository

import (
	"github.com/stretchr/testify/assert"
	"golauth/config/datasource"
	"golauth/model"
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	ctx := Up()
	defer Down(ctx)

	dbTest, err := datasource.CreateDBConnection()
	if err != nil {
		t.Fatal("error when creating datasource: %w", err)
	}

	r := NewUserRepository(dbTest)

	t.Run("load user without password", func(t *testing.T) {
		u, err := r.FindByUsername("admin")
		if err != nil {
			t.Errorf("could not get admin: %s", err.Error())
		}
		assert.NotNil(t, u)
		assert.Equal(t, "admin", u.Username)
		assert.Empty(t, u.Password)
	})

	t.Run("load user with password", func(t *testing.T) {
		u, err := r.FindByUsernameWithPassword("admin")
		if err != nil {
			t.Errorf("could not get admin: %s", err.Error())
		}
		assert.NotNil(t, u)
		assert.Equal(t, "admin", u.Username)
		assert.NotEmpty(t, u.Password)
	})

	t.Run("load user by id", func(t *testing.T) {
		u, err := r.FindByID(1)
		if err != nil {
			t.Errorf("could not user with id 1: %s", err.Error())
		}
		assert.NotNil(t, u)
		assert.Equal(t, "admin", u.Username)
		assert.Empty(t, u.Password)
	})

	t.Run("create new user", func(t *testing.T) {
		u := model.User{
			Username:     "guest",
			FirstName:    "Guest",
			LastName:     "None",
			Email:        "guest@none.com",
			Document:     "123456",
			Password:     "e10adc3949ba59abbe56e057f20f883e",
			Enabled:      true,
			CreationDate: time.Now(),
		}

		user, err := r.Create(u)
		if err != nil {
			t.Errorf("could not create user: %s", err.Error())
		}
		assert.NotEmpty(t, user.ID)
	})
}
