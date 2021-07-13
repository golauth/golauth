//go:generate mockgen -source userRepository.go -destination mock/userRepository_mock.go -package mock
package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"golauth/entity"
)

type UserRepository interface {
	FindByUsername(username string) (entity.User, error)
	FindByUsernameWithPassword(username string) (entity.User, error)
	FindByID(id uuid.UUID) (entity.User, error)
	Create(user entity.User) (entity.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return userRepository{db: db}
}

func (ur userRepository) FindByUsername(username string) (entity.User, error) {
	return ur.findByUsername(username, true)
}

func (ur userRepository) FindByUsernameWithPassword(username string) (entity.User, error) {
	return ur.findByUsername(username, false)
}

func (ur userRepository) findByUsername(username string, omitPassword bool) (entity.User, error) {
	user := entity.User{}
	row := ur.db.QueryRow("SELECT * FROM golauth_user WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Document, &user.Password, &user.Enabled, &user.CreationDate)
	if err != nil {
		return entity.User{}, fmt.Errorf("could not find user by username [%s]: %w", username, err)
	}
	if omitPassword {
		user.Password = ""
	}
	return user, nil
}

func (ur userRepository) FindByID(id uuid.UUID) (entity.User, error) {
	user := entity.User{}
	var phantomZone string
	row := ur.db.QueryRow("SELECT * FROM golauth_user WHERE id = $1", id)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Document, &phantomZone, &user.Enabled, &user.CreationDate)
	if err != nil {
		return entity.User{}, fmt.Errorf("could not find user by id [%d]: %w", id, err)
	}
	return user, nil
}

func (ur userRepository) Create(user entity.User) (entity.User, error) {
	err := ur.db.QueryRow("INSERT INTO golauth_user (username, first_name, last_name, email, document, password) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;",
		user.Username, user.FirstName, user.LastName, user.Email, user.Document, user.Password).Scan(&user.ID)
	if err != nil {
		return entity.User{}, fmt.Errorf("could not create user %s: %w", user.Username, err)
	}
	return user, nil
}
