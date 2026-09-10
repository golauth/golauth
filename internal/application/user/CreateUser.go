//go:generate mockgen -source CreateUser.go -destination mock/CreateUser_mock.go -package mock
package user

import (
	"context"
	"fmt"

	"github.com/golauth/golauth/internal/application/audit"
	"github.com/golauth/golauth/internal/domain/entity"
	"github.com/golauth/golauth/internal/domain/factory"
	"github.com/golauth/golauth/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

const defaultRoleName = "USER"

var bcryptDefaultCost = bcrypt.DefaultCost

type CreateUser interface {
	Execute(ctx context.Context, input *entity.User) (*entity.User, error)
}

func NewCreateUser(repoFactory factory.RepositoryFactory, passwordDenylist PasswordDenylist) CreateUser {
	return createUser{
		userRepository:     repoFactory.NewUserRepository(),
		roleRepository:     repoFactory.NewRoleRepository(),
		userRoleRepository: repoFactory.NewUserRoleRepository(),
		passwordDenylist:   passwordDenylist,
	}
}

type createUser struct {
	userRepository     repository.UserRepository
	roleRepository     repository.RoleRepository
	userRoleRepository repository.UserRoleRepository
	passwordDenylist   PasswordDenylist
}

func (uc createUser) Execute(ctx context.Context, input *entity.User) (*entity.User, error) {
	// Enforce the input policy before touching bcrypt or the database: an
	// invalid payload is a 400, and this is the one place every transport
	// reaches. validateAndNormalize also lower-cases username and email and
	// trims the surrounding fields so the persisted row is canonical.
	if err := validateAndNormalize(input, uc.passwordDenylist); err != nil {
		return nil, err
	}
	// enabled is set here, never accepted from the caller: activating an
	// account is an administrative act, not a self-service one.
	input.Enabled = true
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptDefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not generate password: %w", err)
	}
	input.Password = string(hash)
	savedUser, err := uc.userRepository.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("could not save user: %w", err)
	}
	role, err := uc.roleRepository.FindByName(ctx, defaultRoleName)
	if err != nil {
		return nil, fmt.Errorf("could not fetch default role: %w", err)
	}
	err = uc.userRoleRepository.AddUserRole(ctx, savedUser.ID, role.ID)
	if err != nil {
		return nil, fmt.Errorf("could not add default role to user: %w", err)
	}

	audit.Event(ctx, audit.UserCreated,
		"user_id", savedUser.ID.String(),
		"username", savedUser.Username,
	)
	return savedUser, nil
}
