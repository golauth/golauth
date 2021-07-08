//go:generate mockgen -source signupService.go -destination mock/signupService_mock.go -package mock
package usecase

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golauth/model"
	"golauth/repository"
)

const defaultRoleName = "USER"

var bcryptDefaultCost = bcrypt.DefaultCost

type SignupService interface {
	CreateUser(user model.User) (model.User, error)
}

type signupService struct {
	userRepository     repository.UserRepository
	roleRepository     repository.RoleRepository
	userRoleRepository repository.UserRoleRepository
}

func NewSignupService(
	userRepository repository.UserRepository,
	roleRepository repository.RoleRepository,
	userRoleRepository repository.UserRoleRepository) SignupService {
	return signupService{
		userRepository:     userRepository,
		roleRepository:     roleRepository,
		userRoleRepository: userRoleRepository,
	}
}

func (s signupService) CreateUser(user model.User) (model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcryptDefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("could not generate password: %w", err)
	}

	user.Password = string(hash)
	savedUser, err := s.userRepository.Create(user)
	if err != nil {
		return model.User{}, fmt.Errorf("could not save user: %w", err)
	}
	role, err := s.roleRepository.FindByName(defaultRoleName)
	if err != nil {
		return model.User{}, fmt.Errorf("could not fetch default role: %w", err)
	}
	_, err = s.userRoleRepository.AddUserRole(savedUser.ID, role.ID)
	if err != nil {
		return model.User{}, fmt.Errorf("could not add default role to user: %w", err)
	}

	return savedUser, nil
}
