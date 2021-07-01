package repository

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golauth/config/datasource"
	"golauth/model"
	"testing"
	"time"
)

func TestRoleRepository(t *testing.T) {
	ctx := Up(true)
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
	ctx := Up(true)
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
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows affected")
	})
}

func TestNoSchemaInit(t *testing.T) {
	ctx := Up(false)
	defer Down(ctx)

	dbTest, err := datasource.CreateDBConnection()
	if err != nil {
		t.Fatal("error when creating datasource: %w", err)
	}

	rr := NewRoleRepository(dbTest)

	t.Run("could not find role USER", func(t *testing.T) {
		role, err := rr.FindByName("USER")
		assert.Empty(t, role)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "could not find role USER")
	})

	t.Run("could not create role", func(t *testing.T) {
		r := model.Role{
			Name:         "CUSTOMER_EDIT",
			Description:  "Customer edit",
			Enabled:      true,
			CreationDate: time.Now(),
		}
		role, err := rr.Create(r)
		assert.Empty(t, role)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "could not create role CUSTOMER_EDIT")
	})
}

func TestRoleRepositoryWithMock(t *testing.T) {
	dbTest, mock := newDBMock()
	roleMock := model.Role{Name: "role", Description: "role", Enabled: true}

	rr := NewRoleRepository(dbTest)

	t.Run("FindByName scan error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").
			WithArgs("role").
			WillReturnError(mockScanError)
		result, err := rr.FindByName("role")
		assert.Empty(t, result)
		assert.NotNil(t, err)
		assert.ErrorAs(t, err, &mockScanError)
		assert.Contains(t, err.Error(), "could not find role role")
	})

	t.Run("Create scan error", func(t *testing.T) {
		mock.ExpectQuery("INSERT").
			WithArgs(roleMock.Name, roleMock.Description, roleMock.Enabled).
			WillReturnError(mockScanError)
		result, err := rr.Create(roleMock)
		assert.Empty(t, result)
		assert.NotNil(t, err)
		assert.ErrorAs(t, err, &mockScanError)
		assert.Contains(t, err.Error(), "could not create role role")
	})

	t.Run("Edit exec error", func(t *testing.T) {
		mock.ExpectExec("UPDATE").
			WithArgs(roleMock.ID, roleMock.Name, roleMock.Description, roleMock.Enabled).
			WillReturnError(mockUpdateError)
		err := rr.Edit(roleMock)
		assert.NotNil(t, err)
		assert.ErrorAs(t, err, &mockUpdateError)
		assert.Contains(t, err.Error(), "could not edit role role")
	})

	t.Run("Edit no rows affected", func(t *testing.T) {
		mock.ExpectExec("UPDATE").
			WithArgs(roleMock.ID, roleMock.Name, roleMock.Description, roleMock.Enabled).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err := rr.Edit(roleMock)
		assert.NotNil(t, err)
		assert.ErrorAs(t, err, &sql.ErrNoRows)
		assert.Contains(t, err.Error(), "no rows affected")
	})
}
