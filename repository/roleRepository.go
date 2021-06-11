package repository

import (
	"database/sql"
	"golauth/model"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return RoleRepository{db: db}
}

func (rr RoleRepository) FindByName(name string) (model.Role, error) {
	role := model.Role{}
	row := rr.db.QueryRow("SELECT * FROM golauth_role WHERE name = $1", name)
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.Enabled, &role.CreationDate)
	return role, err
}

func (rr RoleRepository) Create(role model.Role) (model.Role, error) {
	err := rr.db.QueryRow("INSERT INTO golauth_role (name,description,enabled) VALUES ($1, $2, $3) RETURNING id, creation_date;",
		role.Name, role.Description, role.Enabled).Scan(&role.ID, &role.CreationDate)
	return role, err
}

func (rr RoleRepository) Edit(role model.Role) error {
	updateStatement := `
		UPDATE golauth_role
		SET name = $2, description = $3, enabled = $4
		WHERE id = $1
	`
	r, err := rr.db.Exec(updateStatement, role.ID, role.Name, role.Description, role.Enabled)

	nRows, err := r.RowsAffected()
	if nRows == 0 && err == nil {
		err = sql.ErrNoRows
	}
	return err
}
