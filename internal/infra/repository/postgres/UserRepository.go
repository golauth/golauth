package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/domain/entity"
	"github.com/golauth/golauth/internal/domain/repository"
	"github.com/golauth/golauth/internal/infra/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// pgUniqueViolation is the SQLSTATE code Postgres reports for a unique-index
// conflict. Matching the code is stable; matching the message text is not.
const pgUniqueViolation = "23505"

// adminAuthorityName is the authority a user must ultimately hold to administer
// the service.
const adminAuthorityName = "ADMIN"

// userColumns is the full golauth_user column list, in the order every scan in
// this file expects. Naming the columns keeps a scan safe when a migration adds
// or reorders one. userColumnsNoHash is the same list without the hash, for
// callers that must never read it.
const (
	userColumns       = "id, username, first_name, last_name, email, document, password, enabled, creation_date"
	userColumnsNoHash = "id, username, first_name, last_name, email, document, enabled, creation_date"
)

type UserRepositoryPostgres struct {
	db database.Database
}

func NewUserRepository(db database.Database) repository.UserRepository {
	return &UserRepositoryPostgres{db: db}
}

func (ur UserRepositoryPostgres) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	row := ur.db.One(ctx, "SELECT "+userColumns+" FROM golauth_user WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Document, &user.Password, &user.Enabled, &user.CreationDate)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("user %q: %w", username, apperr.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("could not find user by username [%s]: %w", username, err)
	}
	return &user, nil
}

func (ur UserRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	// The password hash is deliberately not selected: no caller of FindByID
	// needs it, and not fetching it is one fewer place it can leak.
	row := ur.db.One(ctx, "SELECT "+userColumnsNoHash+" FROM golauth_user WHERE id = $1", id)
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Document, &user.Enabled, &user.CreationDate)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("user %s: %w", id, apperr.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("could not find user by id [%s]: %w", id, err)
	}
	return &user, nil
}

// AdminExists reports whether any enabled user holds the ADMIN authority
// through an enabled role. A disabled user or a disabled role does not count:
// the service would still be without a usable administrator.
func (ur UserRepositoryPostgres) AdminExists(ctx context.Context) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM golauth_user_role ur
			JOIN golauth_user u            ON u.id = ur.user_id
			JOIN golauth_role r            ON r.id = ur.role_id
			JOIN golauth_role_authority ra ON ra.role_id = r.id
			JOIN golauth_authority a       ON a.id = ra.authority_id
			WHERE a.name = $1 AND ur.enabled AND u.enabled AND r.enabled AND a.enabled
		)`
	var exists bool
	if err := ur.db.One(ctx, q, adminAuthorityName).Scan(&exists); err != nil {
		return false, fmt.Errorf("could not check for an administrator: %w", err)
	}
	return exists, nil
}

func (ur UserRepositoryPostgres) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	err := ur.db.One(ctx, "INSERT INTO golauth_user (username, first_name, last_name, email, document, password) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;",
		user.Username, user.FirstName, user.LastName, user.Email, user.Document, user.Password).Scan(&user.ID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == pgUniqueViolation {
			// The unique index on username / email is the real guard against a
			// duplicate; a pre-check SELECT would only add a race. Surface it as
			// a domain error the controller renders as 409.
			return nil, fmt.Errorf("%s: %w", user.Username, repository.ErrUserAlreadyExists)
		}
		return nil, fmt.Errorf("could not create user %s: %w", user.Username, err)
	}
	return user, nil
}
