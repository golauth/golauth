//go:generate mockgen -source GenerateToken.go -destination mock/GenerateToken_mock.go -package mock
package token

import (
	"context"
	"errors"
	"fmt"
	"github.com/golauth/golauth/domain/factory"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/api/controller/model"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrGeneratingToken           = errors.New("error generating token")
)

type GenerateToken interface {
	Execute(ctx context.Context, username string, password string) (model.TokenResponse, error)
}

func NewGenerateToken(repoFactory factory.RepositoryFactory, tokenService UseCase) GenerateToken {
	return generateToken{
		userRepository:          repoFactory.NewUserRepository(),
		roleRepository:          repoFactory.NewRoleRepository(),
		userRoleRepository:      repoFactory.NewUserRoleRepository(),
		userAuthorityRepository: repoFactory.NewUserAuthorityRepository(),
		tokenService:            tokenService,
	}
}

type generateToken struct {
	userRepository          repository.UserRepository
	roleRepository          repository.RoleRepository
	userRoleRepository      repository.UserRoleRepository
	userAuthorityRepository repository.UserAuthorityRepository
	tokenService            UseCase
}

func (uc generateToken) Execute(ctx context.Context, username string, password string) (model.TokenResponse, error) {
	user, err := uc.userRepository.FindByUsername(ctx, username)
	if err != nil {
		return model.TokenResponse{}, ErrInvalidUsernameOrPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return model.TokenResponse{}, ErrInvalidUsernameOrPassword
	}

	authorities, err := uc.userAuthorityRepository.FindAuthoritiesByUserID(ctx, user.ID)
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("error when fetch authorities: %w", err)
	}

	jwtToken, err := uc.tokenService.GenerateJwtToken(user, authorities)
	if err != nil {
		return model.TokenResponse{}, ErrGeneratingToken
	}
	tokenResponse := model.TokenResponse{AccessToken: jwtToken}
	return tokenResponse, nil
}
