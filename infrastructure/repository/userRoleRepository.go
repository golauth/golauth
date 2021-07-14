//go:generate mockgen -source userRoleRepository.go -destination mock/userRoleRepository_mock.go -package mock
package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type UserRoleRepository interface {
	AddUserRole(userId uuid.UUID, roleId uuid.UUID) error
}

type userRoleRepository struct {
	db *sql.DB
}

func NewUserRoleRepository(db *sql.DB) UserRoleRepository {
	return userRoleRepository{db: db}
}

func (urr userRoleRepository) AddUserRole(userId uuid.UUID, roleId uuid.UUID) error {
	_, err := urr.db.Query("INSERT INTO golauth_user_role (user_id,role_id) VALUES ($1, $2) RETURNING creation_date;",
		userId, roleId)
	if err != nil {
		return fmt.Errorf("could not add userrole [%s;%s]: %w", userId, roleId, err)
	}
	return nil
}
