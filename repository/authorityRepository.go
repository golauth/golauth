package repository

import (
	"golauth/config/db"
	"golauth/model"
	"golauth/util"
)

type AuthorityRepository struct{}

func (u UserRepository) FindByUser(uId int) (interface{}, error) {
	user := model.User{}
	row := db.GetDatasource().QueryRow("SELECT * FROM golauth_role_authority ra INNER JOIN golauth_user_role ur on ur. WHERE username = $1", uId)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.Enabled, &user.CreationDate)

	return util.ResultData(user, model.User{}, err)
}
