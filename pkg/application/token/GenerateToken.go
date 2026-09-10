//go:generate mockgen -source GenerateToken.go -destination mock/GenerateToken_mock.go -package mock
package token

import (
	"context"
	"errors"
	"fmt"

	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/factory"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrGeneratingToken           = errors.New("error generating token")
)

type GenerateToken interface {
	Execute(ctx context.Context, username string, password string) (*entity.Token, error)
}

func NewGenerateToken(repoFactory factory.RepositoryFactory, jwtToken GenerateJwtToken) GenerateToken {
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
	jwtToken                GenerateJwtToken
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

	// A deactivated account must not be able to log in. The caller gets the
	// same error as a wrong password on purpose, so the endpoint cannot be
	// used to probe whether an account exists or is disabled; the real reason
	// is only in the log, keyed by user id and never by the password.
	if !user.Enabled {
		logrus.Infof("token request denied: user %s is disabled", user.ID)
		return nil, ErrInvalidUsernameOrPassword
	}

	// FindAuthoritiesByUserID already filters out disabled roles and
	// authorities, so a user whose roles are all disabled comes back with an
	// empty list. That is intended: an empty list makes generateJwtToken omit
	// the "authorities" claim, and every RequireAuthority check then denies
	// the request. A token with no authority can still be minted -- it just
	// cannot reach any authorized route.
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
