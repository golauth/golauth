package token

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v3"
	"github.com/golauth/golauth/api/handler/model"
	"github.com/golauth/golauth/domain/entity"
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

type Service struct {
	signer   jwt.Signer
	verifier jwt.Verifier
}

func NewService() UseCase {
	s := Service{}
	s.prepare()
	return s
}

func (s *Service) prepare() {
	key := s.generatePrivateKey()
	s.signer = s.generateSigner(key)
	s.verifier = s.generateVerifier(key)
}

func (s Service) generateSigner(key *rsa.PrivateKey) jwt.Signer {
	signer, err := jwt.NewSignerRS(keyAlgorithm, key)
	if err != nil {
		panic(errSignerGenerate)
	}
	return signer
}

func (s Service) generateVerifier(key *rsa.PrivateKey) jwt.Verifier {
	verifier, err := jwt.NewVerifierRS(keyAlgorithm, &key.PublicKey)
	if err != nil {
		panic(errVerifierGenerate)
	}
	return verifier
}

func (s *Service) generatePrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Errorf("could not generate private key: %w", err))
	}
	return privateKey
}

func (s Service) ValidateToken(strToken string) error {
	token, err := jwt.ParseAndVerifyString(strToken, s.verifier)
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

func (s Service) ExtractToken(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	if len(authorization) > len("Bearer ") {
		return authorization[7:], nil
	}
	return "", ErrBearerTokenExtract
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
