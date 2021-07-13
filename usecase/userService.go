//go:generate mockgen -source userService.go -destination mock/userService_mock.go -package mock
package usecase

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golauth/entity"
	"golauth/infrastructure/repository"
	"golauth/model"
)

const defaultRoleName = "USER"

var (
	bcryptDefaultCost            = bcrypt.DefaultCost
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrGeneratingToken           = errors.New("error generating token")
)

type UserService interface {
	CreateUser(user entity.User) (entity.User, error)
	GenerateToken(username string, password string) (model.TokenResponse, error)
}

type userService struct {
	userRepository          repository.UserRepository
	roleRepository          repository.RoleRepository
	userRoleRepository      repository.UserRoleRepository
	userAuthorityRepository repository.UserAuthorityRepository
	tokenService            TokenService
}

func NewUserService(
	userRepository repository.UserRepository,
	roleRepository repository.RoleRepository,
	userRoleRepository repository.UserRoleRepository,
	userAuthorityRepository repository.UserAuthorityRepository,
	tokenService TokenService) UserService {
	return userService{
		userRepository:          userRepository,
		roleRepository:          roleRepository,
		userRoleRepository:      userRoleRepository,
		userAuthorityRepository: userAuthorityRepository,
		tokenService:            tokenService,
	}
}

func (s userService) CreateUser(user entity.User) (entity.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcryptDefaultCost)
	if err != nil {
		return entity.User{}, fmt.Errorf("could not generate password: %w", err)
	}

	user.Password = string(hash)
	savedUser, err := s.userRepository.Create(user)
	if err != nil {
		return entity.User{}, fmt.Errorf("could not save user: %w", err)
	}
	role, err := s.roleRepository.FindByName(defaultRoleName)
	if err != nil {
		return entity.User{}, fmt.Errorf("could not fetch default role: %w", err)
	}
	_, err = s.userRoleRepository.AddUserRole(savedUser.ID, role.ID)
	if err != nil {
		return entity.User{}, fmt.Errorf("could not add default role to user: %w", err)
	}

	return savedUser, nil
}

func (s userService) GenerateToken(username string, password string) (model.TokenResponse, error) {

	user, err := s.userRepository.FindByUsernameWithPassword(username)
	if err != nil {
		return model.TokenResponse{}, ErrInvalidUsernameOrPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return model.TokenResponse{}, ErrInvalidUsernameOrPassword
	}

	authorities, err := s.userAuthorityRepository.FindAuthoritiesByUserID(user.ID)
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("error when fetch authorities: %w", err)
	}

	jwtToken, err := s.tokenService.GenerateJwtToken(user, authorities)
	if err != nil {
		return model.TokenResponse{}, ErrGeneratingToken
	}
	tokenResponse := model.TokenResponse{AccessToken: jwtToken}
	return tokenResponse, nil
}
