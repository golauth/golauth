//go:generate mockgen -source tokenService.go -destination mock/tokenService_mock.go -package mock
package usecase

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v3"
	"golauth/entity"
	"golauth/model"
	"net/http"
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

type TokenService interface {
	ValidateToken(token string) error
	ExtractToken(r *http.Request) (string, error)
	GenerateJwtToken(user entity.User, authorities []string) (string, error)
}

type tokenService struct {
	signer   jwt.Signer
	verifier jwt.Verifier
}

func NewTokenService() TokenService {
	ts := tokenService{}
	ts.prepare()
	return ts
}

func (ts *tokenService) prepare() {
	key := ts.generatePrivateKey()
	ts.signer = ts.generateSigner(key)
	ts.verifier = ts.generateVerifier(key)
}

func (ts tokenService) generateSigner(key *rsa.PrivateKey) jwt.Signer {
	signer, err := jwt.NewSignerRS(keyAlgorithm, key)
	if err != nil {
		panic(errSignerGenerate)
	}
	return signer
}

func (ts tokenService) generateVerifier(key *rsa.PrivateKey) jwt.Verifier {
	verifier, err := jwt.NewVerifierRS(keyAlgorithm, &key.PublicKey)
	if err != nil {
		panic(errVerifierGenerate)
	}
	return verifier
}

func (ts *tokenService) generatePrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Errorf("could not generate private key: %w", err))
	}
	return privateKey
}

func (ts tokenService) ValidateToken(strToken string) error {
	token, err := jwt.ParseAndVerifyString(strToken, ts.verifier)
	if err != nil {
		return fmt.Errorf("could not parse and verify strToken: %w", err)
	}

	claims := &model.Claims{}
	err = json.Unmarshal(token.RawClaims(), &claims)
	if err != nil {
		return fmt.Errorf("could not unmarshal claims: %w", err)
	}
	if !claims.IsValidAt(time.Now()) {
		return errExpiredToken
	}

	return nil
}

func (ts tokenService) ExtractToken(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	if len(authorization) > len("Bearer ") {
		return authorization[7:], nil
	}
	return "", ErrBearerTokenExtract
}

func (ts tokenService) GenerateJwtToken(user entity.User, authorities []string) (string, error) {
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
	builder := jwt.NewBuilder(ts.signer)
	token, err := builder.Build(claims)
	if err != nil {
		return "", fmt.Errorf("could not build token with claims: %w", err)
	}

	return token.String(), nil
}
