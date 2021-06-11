package repository

import (
	"database/sql"
	"golauth/model"
)

type UserRoleRepository struct {
	db *sql.DB
}

func NewUserRoleRepository(db *sql.DB) UserRoleRepository {
	return UserRoleRepository{db: db}
}

func (urr UserRoleRepository) AddUserRole(userId int, roleId int) (model.UserRole, error) {
	userRole := model.UserRole{UserID: userId, RoleID: roleId}
	err := urr.db.QueryRow("INSERT INTO golauth_user_role (user_id,role_id) VALUES ($1, $2) RETURNING creation_date;",
		userRole.UserID, userRole.RoleID).Scan(&userRole.CreationDate)
	return userRole, err
}
