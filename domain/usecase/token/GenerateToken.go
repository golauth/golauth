//go:generate mockgen -source GenerateToken.go -destination mock/GenerateToken_mock.go -package mock
package token

import (
	"context"
	"errors"
	"fmt"
	"github.com/golauth/golauth/core/util"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/factory"
	"github.com/golauth/golauth/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrGeneratingToken           = errors.New("error generating token")
)

type GenerateToken interface {
	Execute(ctx context.Context, username string, password string) (*entity.Token, error)
}

func NewGenerateToken(repoFactory factory.RepositoryFactory, jwtToken util.GenerateJwtToken) GenerateToken {
	return generateToken{
		userRepository:          repoFactory.NewUserRepository(),
		roleRepository:          repoFactory.NewRoleRepository(),
		userRoleRepository:      repoFactory.NewUserRoleRepository(),
		userAuthorityRepository: repoFactory.NewUserAuthorityRepository(),
		jwtToken:                jwtToken,
	}
}

type generateToken struct {
	userRepository          repository.UserRepository
	roleRepository          repository.RoleRepository
	userRoleRepository      repository.UserRoleRepository
	userAuthorityRepository repository.UserAuthorityRepository
	jwtToken                util.GenerateJwtToken
}

func (uc generateToken) Execute(ctx context.Context, username string, password string) (*entity.Token, error) {
	user, err := uc.userRepository.FindByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	authorities, err := uc.userAuthorityRepository.FindAuthoritiesByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error when fetch authorities: %w", err)
	}

	accessToken, err := uc.jwtToken.Execute(user, authorities)
	if err != nil {
		return nil, ErrGeneratingToken
	}
	return &entity.Token{AccessToken: accessToken}, nil
}
