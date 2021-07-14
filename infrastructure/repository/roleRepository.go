//go:generate mockgen -source roleRepository.go -destination mock/roleRepository_mock.go -package mock
package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"golauth/entity"
)

type RoleRepository interface {
	FindByName(name string) (entity.Role, error)
	Create(role entity.Role) (entity.Role, error)
	Edit(role entity.Role) error
	ChangeStatus(id uuid.UUID, enabled bool) error
	ExistsById(id uuid.UUID) (bool, error)
}

type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return roleRepository{db: db}
}

func (r roleRepository) FindByName(name string) (entity.Role, error) {
	role := entity.Role{}
	row := r.db.QueryRow("SELECT * FROM golauth_role WHERE name = $1", name)
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.Enabled, &role.CreationDate)
	if err != nil {
		return entity.Role{}, fmt.Errorf("could not find role %s: %w", name, err)
	}
	return role, nil
}

func (r roleRepository) Create(role entity.Role) (entity.Role, error) {
	err := r.db.QueryRow("INSERT INTO golauth_role (name,description,enabled) VALUES ($1, $2, $3) RETURNING id, creation_date;",
		role.Name, role.Description, role.Enabled).Scan(&role.ID, &role.CreationDate)
	if err != nil {
		return entity.Role{}, fmt.Errorf("could not create role %s: %w", role.Name, err)
	}
	return role, nil
}

func (r roleRepository) Edit(role entity.Role) error {
	updateStatement := `
		UPDATE golauth_role
		SET name = $2, description = $3
		WHERE id = $1
	`
	res, err := r.db.Exec(updateStatement, role.ID, role.Name, role.Description)
	if err != nil {
		return fmt.Errorf("could not edit role %s: %w", role.Name, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("no rows affected: %w", err)
	}
	return nil
}

func (r roleRepository) ChangeStatus(id uuid.UUID, enabled bool) error {
	updateStatement := `
		UPDATE golauth_role
		SET enabled = $2
		WHERE id = $1
	`
	res, err := r.db.Exec(updateStatement, id, enabled)
	if err != nil {
		return fmt.Errorf("could not edit role %s: %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("no rows affected: %w", err)
	}
	return nil
}

func (r roleRepository) ExistsById(id uuid.UUID) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM golauth_role WHERE id = $1)"
	row := r.db.QueryRow(query, id)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
