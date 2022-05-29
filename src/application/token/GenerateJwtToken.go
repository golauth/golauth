//go:generate mockgen -source GenerateJwtToken.go -destination mock/GenerateJwtToken_mock.go -package mock
package token

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v3"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/infra/api/controller/model"
	"time"
)

var (
	ErrBearerTokenExtract = errors.New("bearer token extract error")
	errSignerGenerate     = errors.New("could not generate signer from private key")
	errVerifierGenerate   = errors.New("could not generate verifier from public key")
	keyAlgorithm          = jwt.RS512
	TokenExpirationTime   = 60
)

type GenerateJwtToken interface {
	Execute(user *entity.User, authorities []string) (string, error)
}

func NewGenerateJwtToken(key *rsa.PrivateKey) GenerateJwtToken {
	return generateJwtToken{signer: GenerateSigner(key)}
}

type generateJwtToken struct {
	signer jwt.Signer
}

func (uc generateJwtToken) Execute(user *entity.User, authorities []string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(TokenExpirationTime) * time.Minute)
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
	tk, err := builder.Build(claims)
	if err != nil {
		return "", fmt.Errorf("could not build token with claims: %w", err)
	}

	return tk.String(), nil
}

func GenerateSigner(key *rsa.PrivateKey) jwt.Signer {
	signer, err := jwt.NewSignerRS(keyAlgorithm, key)
	if err != nil {
		panic(errSignerGenerate)
	}
	return signer
}

func GeneratePrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Errorf("could not generate private key: %w", err))
	}
	return privateKey
}

func GenerateVerifier(key *rsa.PrivateKey) jwt.Verifier {
	verifier, err := jwt.NewVerifierRS(keyAlgorithm, &key.PublicKey)
	if err != nil {
		panic(errVerifierGenerate)
	}
	return verifier
}
