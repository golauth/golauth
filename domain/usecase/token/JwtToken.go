//go:generate mockgen -source JwtToken.go -destination mock/JwtToken_mock.go -package mock
package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v3"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/infra/api/controller/model"
	"time"
)

var (
	ErrBearerTokenExtract = errors.New("bearer token extract error")
	errExpiredToken       = errors.New("expired token")
	errSignerGenerate     = errors.New("could not generate signer from private key")
	errVerifierGenerate   = errors.New("could not generate verifier from public key")
	keyAlgorithm          = jwt.RS512
	tokenExpirationTime   = 30
)

type JwtToken interface {
	Execute(user entity.User, authorities []string) (string, error)
}

func NewJwtToken(key *rsa.PrivateKey) JwtToken {
	return jwtToken{signer: generateSigner(key)}
}

type jwtToken struct {
	signer jwt.Signer
}

func (uc jwtToken) Execute(user entity.User, authorities []string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(tokenExpirationTime) * time.Minute)
	claims := &model.Claims{
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Authorities: authorities,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	builder := jwt.NewBuilder(uc.signer)
	token, err := builder.Build(claims)
	if err != nil {
		return "", fmt.Errorf("could not build token with claims: %w", err)
	}

	return token.String(), nil
}
