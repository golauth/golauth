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

type Service struct {
	signer jwt.Signer
}

func NewService(key *rsa.PrivateKey) UseCase {
	return Service{signer: generateSigner(key)}
}

func (s Service) GenerateJwtToken(user entity.User, authorities []string) (string, error) {
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
	builder := jwt.NewBuilder(s.signer)
	token, err := builder.Build(claims)
	if err != nil {
		return "", fmt.Errorf("could not build token with claims: %w", err)
	}

	return token.String(), nil
}
