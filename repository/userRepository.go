package repository

import (
	"database/sql"
	"golauth/model"
)

type UserRepository interface {
	FindByUsername(username string) (model.User, error)
	FindByUsernameWithPassword(username string) (model.User, error)
	FindByID(id int) (model.User, error)
	Create(user model.User) (model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return userRepository{db: db}
}

func (ur userRepository) FindByUsername(username string) (model.User, error) {
	return ur.findByUsername(username, true)
}

func (ur userRepository) FindByUsernameWithPassword(username string) (model.User, error) {
	return ur.findByUsername(username, false)
}

func (ur userRepository) findByUsername(username string, omitPassword bool) (model.User, error) {
	user := model.User{}
	row := ur.db.QueryRow("SELECT * FROM golauth_user WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Document, &user.Password, &user.Enabled, &user.CreationDate)
	if omitPassword {
		user.Password = ""
	}
	return user, err
}

func (ur userRepository) FindByID(id int) (model.User, error) {
	user := model.User{}
	var phantomZone string
	row := ur.db.QueryRow("SELECT * FROM golauth_user WHERE id = $1", id)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Document, &phantomZone, &user.Enabled, &user.CreationDate)
	return user, err
}

func (ur userRepository) Create(user model.User) (model.User, error) {
	err := ur.db.QueryRow("INSERT INTO golauth_user (username, first_name, last_name, email, document, password) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;",
		user.Username, user.FirstName, user.LastName, user.Email, user.Document, user.Password).Scan(&user.ID)
	return user, err
}
