package repository

import (
	"database/sql"
	"fmt"
)

type UserAuthorityRepository interface {
	FindAuthoritiesByUserID(userId int) ([]string, error)
}

type userAuthorityRepository struct {
	db *sql.DB
}

func NewUserAuthorityRepository(db *sql.DB) UserAuthorityRepository {
	return userAuthorityRepository{db: db}
}

func (u userAuthorityRepository) FindAuthoritiesByUserID(userId int) ([]string, error) {
	var authorities []string
	var err error
	var query = `
		SELECT a.name 
		FROM golauth_authority a 
		    INNER JOIN golauth_role_authority ra ON ra.authority_id = a.id 
		    INNER JOIN golauth_user_role ur ON ur.role_id = ra.role_id 
		WHERE ur.user_id = $1`

	rows, err := u.db.Query(query, userId)
	if err != nil {
		return nil, fmt.Errorf("could not find authorities by user: %w", err)
	}

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, fmt.Errorf("could not transform result in slice: %w", err)
		}
		authorities = append(authorities, name)
	}

	return authorities, nil
}
