package postgres

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/domain/repository"
	"github.com/golauth/golauth/src/infra/database"
	"github.com/google/uuid"
)

type UserRepositoryPostgres struct {
	db database.Database
}

func NewUserRepository(db database.Database) repository.UserRepository {
	return &UserRepositoryPostgres{db: db}
}

func (ur UserRepositoryPostgres) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	row := ur.db.One(ctx, "SELECT * FROM golauth_user WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Document, &user.Password, &user.Enabled, &user.CreationDate)
	if err != nil {
		return nil, fmt.Errorf("could not find user by username [%s]: %w", username, err)
	}
	return &user, nil
}

func (ur UserRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	var phantomZone string
	row := ur.db.One(ctx, "SELECT * FROM golauth_user WHERE id = $1", id)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Document, &phantomZone, &user.Enabled, &user.CreationDate)
	if err != nil {
		return nil, fmt.Errorf("could not find user by id [%d]: %w", id, err)
	}
	return &user, nil
}

func (ur UserRepositoryPostgres) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	err := ur.db.One(ctx, "INSERT INTO golauth_user (username, first_name, last_name, email, document, password) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;",
		user.Username, user.FirstName, user.LastName, user.Email, user.Document, user.Password).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", user.Username, err)
	}
	return user, nil
}
