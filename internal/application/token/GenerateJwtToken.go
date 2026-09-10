//go:generate mockgen -source GenerateJwtToken.go -destination mock/GenerateJwtToken_mock.go -package mock
package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/golauth/golauth/internal/application/keys"
	"github.com/golauth/golauth/internal/application/token/claims"
	"github.com/golauth/golauth/internal/domain/entity"
)

var (
	ErrBearerTokenExtract = errors.New("bearer token extract error")
	errSignerGenerate     = errors.New("could not generate signer from private key")
	errVerifierGenerate   = errors.New("could not generate verifier from public key")
	keyAlgorithm          = jwt.RS512
)

type GenerateJwtToken interface {
	Execute(user *entity.User, authorities []string) (string, error)
}

// NewGenerateJwtToken builds the access-token signer. accessTTL is the token
// lifetime; a non-positive value falls back to DefaultAccessTokenTTL so a
// misconfiguration cannot mint tokens that never (or instantly) expire.
func NewGenerateJwtToken(key *keys.SigningKey, accessTTL time.Duration) GenerateJwtToken {
	if accessTTL <= 0 {
		accessTTL = DefaultAccessTokenTTL
	}
	return generateJwtToken{signer: GenerateSigner(key.Private), kid: key.KID, accessTTL: accessTTL}
}

type generateJwtToken struct {
	signer    jwt.Signer
	kid       string
	accessTTL time.Duration
}

func (uc generateJwtToken) Execute(user *entity.User, authorities []string) (string, error) {
	expirationTime := time.Now().Add(uc.accessTTL)
	c := &claims.Claims{
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
	tk, err := builder.Build(c)
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
