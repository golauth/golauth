package repository

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"golauth/config/datasource"
	"golauth/model"
	"testing"
	"time"
)

func TestRoleRepository(t *testing.T) {
	ctx := Up()
	defer Down(ctx)

	dbTest, err := datasource.CreateDBConnection()
	if err != nil {
		t.Fatal("error when creating datasource: %w", err)
	}

	rr := NewRoleRepository(dbTest)

	t.Run("find role by name", func(t *testing.T) {
		role, err := rr.FindByName("USER")
		if err != nil {
			t.Errorf("could not find role by name: %s", err.Error())
		}
		assert.NotNil(t, role)
		assert.Equal(t, "USER", role.Name)
	})

	t.Run("create new role", func(t *testing.T) {
		r := model.Role{
			Name:         "CUSTOMER_EDIT",
			Description:  "Customer edit",
			Enabled:      true,
			CreationDate: time.Now(),
		}
		role, err := rr.Create(r)
		if err != nil {
			t.Errorf("could not create role: %s", err.Error())
		}
		assert.NotNil(t, role)
		assert.NotNil(t, role.ID)
		assert.Equal(t, "CUSTOMER_EDIT", role.Name)
	})
}

func TestRoleRepository_Edit(t *testing.T) {
	ctx := Up()
	defer Down(ctx)

	dbTest, err := datasource.CreateDBConnection()
	if err != nil {
		t.Fatal("error when creating datasource: %w", err)
	}

	rr := NewRoleRepository(dbTest)

	t.Run("edit success", func(t *testing.T) {

		r, err := rr.FindByName("USER")
		if err != nil {
			t.Errorf("could not find role by name: %s", err.Error())
		}
		assert.NotNil(t, r)
		assert.Equal(t, "Role USER", r.Description)

		r.Description = "Role to common user"
		err = rr.Edit(r)
		if err != nil {
			t.Errorf("could not edit role: %s", err.Error())
		}

		edited, err := rr.FindByName("USER")
		if err != nil {
			t.Errorf("could not find role by name: %s", err.Error())
		}
		assert.NotNil(t, edited)
		assert.Equal(t, "Role to common user", edited.Description)
	})

	t.Run("edit id not found", func(t *testing.T) {
		r := model.Role{
			ID:           999,
			Name:         "CUSTOMER_EDIT",
			Description:  "Customer edit",
			Enabled:      true,
			CreationDate: time.Now(),
		}
		err := rr.Edit(r)
		assert.Equal(t, sql.ErrNoRows, err)
	})
}
