package repository

import (
	"golauth/config/db"
	"golauth/model"
	"golauth/util"
)

type UserRepository struct{}

func (u UserRepository) FindByUsername(username string) (interface{}, error) {
	return u.findByUsername(username, true)
}

func (u UserRepository) FindByUsernameWithPassword(username string) (interface{}, error) {
	return u.findByUsername(username, false)
}

func (u UserRepository) findByUsername(username string, omitPassword bool) (interface{}, error) {
	user := model.User{}
	row := db.GetDatasource().QueryRow("SELECT * FROM golauth_user WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.Enabled, &user.CreationDate)
	if omitPassword {
		user.Password = ""
	}
	return util.ResultData(user, model.User{}, err)
}

func (u UserRepository) FindByID(id int) (interface{}, error) {
	user := model.User{}
	var phantonZone string
	row := db.GetDatasource().QueryRow("SELECT * FROM golauth_user WHERE id = $1", id)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &phantonZone, &user.Enabled, &user.CreationDate)
	return util.ResultData(user, model.User{}, err)
}

func (u UserRepository) Create(user model.User) (interface{}, error) {
	err := db.GetDatasource().QueryRow("INSERT INTO golauth_user (username, first_name, last_name, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING id;",
		user.Username, user.FirstName, user.LastName, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		return util.ResultData(nil, nil, err)
	}
	return u.FindByID(user.ID)
}
