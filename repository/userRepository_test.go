package repository

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golauth/config/datasource"
	"golauth/model"
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	ctx := Up(true)
	defer Down(ctx)

	ds, err := datasource.NewDatasource()
	assert.NoError(t, err)
	dbTest := ds.GetDB()

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

func TestUserRepositoryWithMock(t *testing.T) {
	dbTest, mock := newDBMock()
	repo := NewUserRepository(dbTest)
	defer func() {
		_ = dbTest.Close()
	}()

	t.Run("FindByUsername scan error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").
			WithArgs("username").
			WillReturnError(mockScanError)
		result, err := repo.FindByUsername("username")
		assert.Empty(t, result)
		assert.NotNil(t, err)
		assert.ErrorAs(t, err, &mockScanError)
		assert.Contains(t, err.Error(), "could not find user by username [username]")
	})

	t.Run("FindByID scan error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").
			WithArgs(1).
			WillReturnError(mockScanError)
		result, err := repo.FindByID(1)
		assert.Empty(t, result)
		assert.NotNil(t, err)
		assert.ErrorAs(t, err, &mockScanError)
		assert.Contains(t, err.Error(), "could not find user by id [1]")
	})

	t.Run("Create scan error", func(t *testing.T) {
		mock.ExpectQuery("INSERT").
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(mockScanError)
		result, err := repo.Create(model.User{Username: "username"})
		assert.Empty(t, result)
		assert.NotNil(t, err)
		assert.ErrorAs(t, err, &mockScanError)
		assert.Contains(t, err.Error(), "could not create user username")
	})
}
