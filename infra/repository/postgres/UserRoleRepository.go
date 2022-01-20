package postgres

import (
	"database/sql"
	"fmt"
	"github.com/golauth/golauth/domain/repository"
	"github.com/google/uuid"
)

type UserRoleRepositoryPostgres struct {
	db *sql.DB
}

func NewUserRoleRepository(db *sql.DB) repository.UserRoleRepository {
	return &UserRoleRepositoryPostgres{db: db}
}

func (urr UserRoleRepositoryPostgres) AddUserRole(userId uuid.UUID, roleId uuid.UUID) error {
	_, err := urr.db.Query("INSERT INTO golauth_user_role (user_id,role_id) VALUES ($1, $2) RETURNING creation_date;",
		userId, roleId)
	if err != nil {
		return fmt.Errorf("could not add userrole [%s;%s]: %w", userId, roleId, err)
	}
	return nil
}
