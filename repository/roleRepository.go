package repository

import (
	"golauth/config/db"
	"golauth/model"
)

type RoleRepository struct{}

func (roleRepository RoleRepository) FindByName(name string) (model.Role, error) {
	role := model.Role{}
	row := db.GetDatasource().QueryRow("SELECT * FROM golauth_role WHERE name = $1", name)
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.Enabled, &role.CreationDate)
	return role, err
}
