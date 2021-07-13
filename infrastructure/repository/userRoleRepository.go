//go:generate mockgen -source userRoleRepository.go -destination mock/userRoleRepository_mock.go -package mock
package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"golauth/entity"
)

type UserRoleRepository interface {
	AddUserRole(userId uuid.UUID, roleId uuid.UUID) (entity.UserRole, error)
}

type userRoleRepository struct {
	db *sql.DB
}

func NewUserRoleRepository(db *sql.DB) UserRoleRepository {
	return userRoleRepository{db: db}
}

func (urr userRoleRepository) AddUserRole(userId uuid.UUID, roleId uuid.UUID) (entity.UserRole, error) {
	userRole := entity.UserRole{UserID: userId, RoleID: roleId}
	err := urr.db.QueryRow("INSERT INTO golauth_user_role (user_id,role_id) VALUES ($1, $2) RETURNING creation_date;",
		userRole.UserID, userRole.RoleID).Scan(&userRole.CreationDate)
	if err != nil {
		return entity.UserRole{}, fmt.Errorf("could not add userrole [user:%d:role:%d]: %w", userId, roleId, err)
	}
	return userRole, err
}
