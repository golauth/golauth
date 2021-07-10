//go:generate mockgen -source tokenService.go -destination mock/tokenService_mock.go -package mock
package usecase

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golauth/model"
	"net/http"
	"time"
)

var ErrBearerTokenExtract = errors.New("bearer token extract error")

const tokenExpirationTime = 30

type TokenService interface {
	ValidateToken(token string) error
	ExtractToken(r *http.Request) (string, error)
	GenerateJwtToken(user model.User, authorities []string) (interface{}, error)
}

type tokenService struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func NewTokenService(privBytes []byte, pubBytes []byte) TokenService {
	ts := tokenService{}
	ts.parseKeys(privBytes, pubBytes)
	return ts
}

func (ts *tokenService) parseKeys(privBytes []byte, pubBytes []byte) {
	var err error
	ts.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		panic(fmt.Errorf("could not parse RSA private key from pem: %w", err))
	}

	ts.publicKey, err = jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		panic(fmt.Errorf("could not parse RSA public key from pem: %w", err))
	}
}

func (ts tokenService) ValidateToken(token string) error {
	claims := &model.Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (i interface{}, err error) {
		return ts.publicKey, nil
	})

	if err != nil {
		return fmt.Errorf("error when parse with claims: %w", err)
	}

	err = parsedToken.Claims.Valid()
	if err != nil {
		return fmt.Errorf("parsed token claims invalid: %w", err)
	}

	if !parsedToken.Valid {
		return fmt.Errorf("parsed token invalid")
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

func (ts tokenService) GenerateJwtToken(user model.User, authorities []string) (interface{}, error) {
	expirationTime := time.Now().Add(tokenExpirationTime * time.Minute)
	claims := &model.Claims{
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Authorities: authorities,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(ts.privateKey)
}
