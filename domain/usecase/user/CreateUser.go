//go:generate mockgen -source CreateUser.go -destination mock/CreateUser_mock.go -package mock
package user

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/factory"
	"github.com/golauth/golauth/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

const defaultRoleName = "USER"

var bcryptDefaultCost = bcrypt.DefaultCost

type CreateUser interface {
	Execute(ctx context.Context, input *entity.User) (*entity.User, error)
}

func NewCreateUser(repoFactory factory.RepositoryFactory) CreateUser {
	return createUser{
		userRepository:     repoFactory.NewUserRepository(),
		roleRepository:     repoFactory.NewRoleRepository(),
		userRoleRepository: repoFactory.NewUserRoleRepository(),
	}
}

type createUser struct {
	userRepository     repository.UserRepository
	roleRepository     repository.RoleRepository
	userRoleRepository repository.UserRoleRepository
}

func (uc createUser) Execute(ctx context.Context, input *entity.User) (*entity.User, error) {
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

	return savedUser, nil
}
