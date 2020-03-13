package repository

import (
	"golauth/config/db"
	"golauth/model"
	"golauth/util"
)

type UserRepository struct{}

func (u UserRepository) FindByUsername(username string) (interface{}, error) {
	user := model.User{}
	row := db.GetDatasource().QueryRow("SELECT * FROM golauth_user WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.Enabled, &user.CreationDate)

	return util.ResultData(user, model.User{}, err)
}

func (u UserRepository) Create(user model.User) (interface{}, error) {
	err := db.GetDatasource().QueryRow("INSERT INTO golauth_user (username, first_name, last_name, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING id;",
		user.Username, user.FirstName, user.LastName, user.Email, user.Password).Scan(&user.ID)
	return util.ResultData(user.ID, 0, err)
}
