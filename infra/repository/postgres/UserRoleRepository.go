package postgres

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/database"
	"github.com/google/uuid"
)

type UserRoleRepositoryPostgres struct {
	db database.Database
}

func NewUserRoleRepository(db database.Database) repository.UserRoleRepository {
	return &UserRoleRepositoryPostgres{db: db}
}

func (urr UserRoleRepositoryPostgres) AddUserRole(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error {
	_, err := urr.db.Exec(ctx, "INSERT INTO golauth_user_role (user_id,role_id) VALUES ($1, $2) RETURNING creation_date;",
		userId, roleId)
	if err != nil {
		return fmt.Errorf("could not add userrole [%s;%s]: %w", userId, roleId, err)
	}
	return nil
}
