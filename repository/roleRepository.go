//go:generate mockgen -source roleRepository.go -destination mock/roleRepository_mock.go -package mock
package repository

import (
	"database/sql"
	"fmt"
	"golauth/model"
)

type RoleRepository interface {
	FindByName(name string) (model.Role, error)
	Create(role model.Role) (model.Role, error)
	Edit(role model.Role) error
}

type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return roleRepository{db: db}
}

func (rr roleRepository) FindByName(name string) (model.Role, error) {
	role := model.Role{}
	row := rr.db.QueryRow("SELECT * FROM golauth_role WHERE name = $1", name)
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.Enabled, &role.CreationDate)
	if err != nil {
		return model.Role{}, fmt.Errorf("could not find role %s: %w", name, err)
	}
	return role, nil
}

func (rr roleRepository) Create(role model.Role) (model.Role, error) {
	err := rr.db.QueryRow("INSERT INTO golauth_role (name,description,enabled) VALUES ($1, $2, $3) RETURNING id, creation_date;",
		role.Name, role.Description, role.Enabled).Scan(&role.ID, &role.CreationDate)
	if err != nil {
		return model.Role{}, fmt.Errorf("could not create role %s: %w", role.Name, err)
	}
	return role, nil
}

func (rr roleRepository) Edit(role model.Role) error {
	updateStatement := `
		UPDATE golauth_role
		SET name = $2, description = $3, enabled = $4
		WHERE id = $1
	`
	r, err := rr.db.Exec(updateStatement, role.ID, role.Name, role.Description, role.Enabled)
	if err != nil {
		return fmt.Errorf("could not edit role %s: %w", role.Name, err)
	}

	nRows, err := r.RowsAffected()
	if err != nil || nRows == 0 {
		return fmt.Errorf("no rows affected: %w", err)
	}
	return nil
}
