package postgres

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/google/uuid"
)

type RoleRepositoryPostgres struct {
	db database.Database
}

func NewRoleRepository(db database.Database) repository.RoleRepository {
	return &RoleRepositoryPostgres{db: db}
}

func (r RoleRepositoryPostgres) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	role := entity.Role{}
	row := r.db.One(ctx, "SELECT * FROM golauth_role WHERE name = $1", name)
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.Enabled, &role.CreationDate)
	if err != nil {
		return nil, fmt.Errorf("could not find role %s: %w", name, err)
	}
	return &role, nil
}

func (r RoleRepositoryPostgres) Create(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	err := r.db.One(ctx, "INSERT INTO golauth_role (name,description,enabled) VALUES ($1, $2, $3) RETURNING id, creation_date;",
		role.Name, role.Description, role.Enabled).Scan(&role.ID, &role.CreationDate)
	if err != nil {
		return nil, fmt.Errorf("could not create role %s: %w", role.Name, err)
	}
	return role, nil
}

func (r RoleRepositoryPostgres) Edit(ctx context.Context, role *entity.Role) error {
	updateStatement := `
		UPDATE golauth_role
		SET name = $2, description = $3
		WHERE id = $1
	`
	res, err := r.db.Exec(ctx, updateStatement, role.ID, role.Name, role.Description)
	if err != nil {
		return fmt.Errorf("could not edit role %s: %w", role.Name, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("no rows affected: %w", err)
	}
	return nil
}

func (r RoleRepositoryPostgres) ChangeStatus(ctx context.Context, id uuid.UUID, enabled bool) error {
	updateStatement := `
		UPDATE golauth_role
		SET enabled = $2
		WHERE id = $1
	`
	res, err := r.db.Exec(ctx, updateStatement, id, enabled)
	if err != nil {
		return fmt.Errorf("could not edit role %s: %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("no rows affected: %w", err)
	}
	return nil
}

func (r RoleRepositoryPostgres) ExistsById(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM golauth_role WHERE id = $1)"
	row := r.db.One(ctx, query, id)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
