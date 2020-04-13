package repository

import (
	"golauth/config/db"
	"golauth/model"
	"golauth/util"
)

type UserRoleRepository struct{}

func (urr UserRoleRepository) AddUserRole(userId int, roleId int) (interface{}, error) {
	userRole := model.UserRole{UserID: userId, RoleID: roleId}
	err := db.GetDatasource().QueryRow("INSERT INTO golauth_user_role (user_id,role_id) VALUES ($1, $2) RETURNING creation_date;",
		userRole.UserID, userRole.RoleID).Scan(&userRole.CreationDate)
	return util.ResultData(userRole, model.UserRole{}, err)
}
