//go:generate mockgen -source userService.go -destination mock/userService_mock.go -package mock
package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/domain/usecase/token"
	"github.com/golauth/golauth/infra/api/controller/model"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrGeneratingToken           = errors.New("error generating token")
)

type UserService interface {
	GenerateToken(ctx context.Context, username string, password string) (model.TokenResponse, error)
}

type userService struct {
	userRepository          repository.UserRepository
	roleRepository          repository.RoleRepository
	userRoleRepository      repository.UserRoleRepository
	userAuthorityRepository repository.UserAuthorityRepository
	tokenService            token.UseCase
}

func NewUserService(
	userRepository repository.UserRepository,
	roleRepository repository.RoleRepository,
	userRoleRepository repository.UserRoleRepository,
	userAuthorityRepository repository.UserAuthorityRepository,
	tokenService token.UseCase) UserService {
	return userService{
		userRepository:          userRepository,
		roleRepository:          roleRepository,
		userRoleRepository:      userRoleRepository,
		userAuthorityRepository: userAuthorityRepository,
		tokenService:            tokenService,
	}
}

func (s userService) GenerateToken(ctx context.Context, username string, password string) (model.TokenResponse, error) {
	user, err := s.userRepository.FindByUsername(ctx, username)
	if err != nil {
		return model.TokenResponse{}, ErrInvalidUsernameOrPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return model.TokenResponse{}, ErrInvalidUsernameOrPassword
	}

	authorities, err := s.userAuthorityRepository.FindAuthoritiesByUserID(ctx, user.ID)
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
