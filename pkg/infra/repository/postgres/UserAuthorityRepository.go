package postgres

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/google/uuid"
)

type UserAuthorityRepositoryPostgres struct {
	db database.Database
}

func NewUserAuthorityRepository(db database.Database) repository.UserAuthorityRepository {
	return &UserAuthorityRepositoryPostgres{db: db}
}

func (u UserAuthorityRepositoryPostgres) FindAuthoritiesByUserID(ctx context.Context, userId uuid.UUID) ([]string, error) {
	var authorities []string
	var err error
	// A disabled role or a disabled authority contributes nothing: the
	// enabled flags are the account-deactivation control. golauth_user_role
	// has no enabled column yet; Plan 10 adds it and the predicate ur.enabled
	// belongs in this same WHERE clause once it does.
	var query = `
		SELECT a.name
		FROM golauth_authority a
		    INNER JOIN golauth_role_authority ra ON ra.authority_id = a.id
		    INNER JOIN golauth_user_role ur ON ur.role_id = ra.role_id
		    INNER JOIN golauth_role r ON r.id = ra.role_id
		WHERE ur.user_id = $1
		    AND a.enabled
		    AND r.enabled`

	rows, err := u.db.Many(ctx, query, userId)
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
