package repository

import (
	"database/sql"
	"golauth/util"
)

type UserAuthorityRepository struct {
	db *sql.DB
}

func NewUserAuthorityRepository(db *sql.DB) UserAuthorityRepository {
	return UserAuthorityRepository{db: db}
}

func (u UserAuthorityRepository) FindAuthoritiesByUserID(userId int) ([]string, error) {
	var authorities []string
	var err error
	rows, _ := u.db.Query("SELECT a.name FROM golauth_authority a INNER JOIN golauth_role_authority ra ON ra.authority_id = a.id INNER JOIN golauth_user_role ur ON ur.role_id = ra.role_id WHERE ur.user_id = $1", userId)

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			break
		}
		authorities = append(authorities, name)
	}

	return util.ResultSliceString(authorities, []string{}, err)
}
