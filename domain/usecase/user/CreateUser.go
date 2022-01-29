//go:generate mockgen -source CreateUser.go -destination mock/CreateUser_mock.go -package mock
package user

import (
	"context"
	"fmt"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/factory"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/api/controller/model"
	"golang.org/x/crypto/bcrypt"
)

const defaultRoleName = "USER"

var bcryptDefaultCost = bcrypt.DefaultCost

type CreateUser interface {
	Execute(ctx context.Context, userReq model.UserRequest) (model.UserResponse, error)
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

func (uc createUser) Execute(ctx context.Context, userReq model.UserRequest) (model.UserResponse, error) {
	user := entity.User{
		Username:  userReq.Username,
		FirstName: userReq.FirstName,
		LastName:  userReq.LastName,
		Email:     userReq.Email,
		Document:  userReq.Document,
		Password:  userReq.Password,
		Enabled:   true,
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcryptDefaultCost)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("could not generate password: %w", err)
	}
	user.Password = string(hash)
	savedUser, err := uc.userRepository.Create(ctx, user)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("could not save user: %w", err)
	}
	role, err := uc.roleRepository.FindByName(ctx, defaultRoleName)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("could not fetch default role: %w", err)
	}
	err = uc.userRoleRepository.AddUserRole(ctx, savedUser.ID, role.ID)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("could not add default role to user: %w", err)
	}

	return model.UserResponse{
		ID:           savedUser.ID,
		Username:     savedUser.Username,
		FirstName:    savedUser.FirstName,
		LastName:     savedUser.LastName,
		Email:        savedUser.Email,
		Document:     savedUser.Document,
		Enabled:      savedUser.Enabled,
		CreationDate: savedUser.CreationDate,
	}, nil
}
