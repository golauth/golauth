//go:generate mockgen -source GenerateJwtToken.go -destination mock/GenerateJwtToken_mock.go -package mock
package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/golauth/golauth/pkg/infra/keys"
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

func NewGenerateJwtToken(key *keys.SigningKey) GenerateJwtToken {
	return generateJwtToken{signer: GenerateSigner(key.Private), kid: key.KID}
}

type generateJwtToken struct {
	signer jwt.Signer
	kid    string
}

func (uc generateJwtToken) Execute(user *entity.User, authorities []string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(TokenExpirationTime) * time.Minute)
	claims := &model.Claims{
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Authorities: authorities,
		StandardClaims: jwt.StandardClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	// The kid header lets a verifier -- another replica, or an offline
	// consumer reading the JWKS -- pick the right public key without trying
	// each one.
	builder := jwt.NewBuilder(uc.signer, jwt.WithKeyID(uc.kid))
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

func GenerateVerifier(key *rsa.PrivateKey) jwt.Verifier {
	verifier, err := jwt.NewVerifierRS(keyAlgorithm, &key.PublicKey)
	if err != nil {
		panic(errVerifierGenerate)
	}
	return verifier
}
