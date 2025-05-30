//go:generate mockgen -source ValidateToken.go -destination mock/ValidateToken_mock.go -package mock
package token

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v3"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"time"
)

var errExpiredToken = errors.New("expired token")

type ValidateToken interface {
	Execute(token string) error
}

func NewValidateToken(key *rsa.PrivateKey) ValidateToken {
	return validateToken{verifier: GenerateVerifier(key)}
}

type validateToken struct {
	verifier jwt.Verifier
}

func (uc validateToken) Execute(strToken string) error {
	token, err := jwt.ParseAndVerifyString(strToken, uc.verifier)
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
