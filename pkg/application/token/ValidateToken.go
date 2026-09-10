//go:generate mockgen -source ValidateToken.go -destination mock/ValidateToken_mock.go -package mock
package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/golauth/golauth/pkg/infra/keys"
)

var (
	errExpiredToken = errors.New("expired token")
	errUnknownKeyID = errors.New("no verifier for token key id")
)

type ValidateToken interface {
	Execute(token string) (*model.Claims, error)
}

func NewValidateToken(ks *keys.KeySet) ValidateToken {
	verifiers := make(map[string]jwt.Verifier, len(ks.Previous)+1)
	for _, sk := range ks.All() {
		verifiers[sk.KID] = GenerateVerifier(sk.Private)
	}
	return validateToken{
		verifiers: verifiers,
		// fallback verifies tokens minted before kid headers existed; they
		// can only have been signed by the current key.
		fallback: GenerateVerifier(ks.Current.Private),
	}
}

type validateToken struct {
	verifiers map[string]jwt.Verifier
	fallback  jwt.Verifier
}

func (uc validateToken) Execute(strToken string) (*model.Claims, error) {
	token, err := jwt.ParseString(strToken)
	if err != nil {
		return nil, fmt.Errorf("could not parse and verify strToken: %w", err)
	}
	if token.Header().Algorithm != keyAlgorithm {
		return nil, fmt.Errorf("could not parse and verify strToken: %w", jwt.ErrAlgorithmMismatch)
	}

	verifier := uc.fallback
	if kid := token.Header().KeyID; kid != "" {
		v, ok := uc.verifiers[kid]
		if !ok {
			return nil, errUnknownKeyID
		}
		verifier = v
	}
	if err = verifier.Verify(token.Payload(), token.Signature()); err != nil {
		return nil, fmt.Errorf("could not parse and verify strToken: %w", err)
	}

	claims := &model.Claims{}
	err = json.Unmarshal(token.RawClaims(), &claims)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal claims: %w", err)
	}
	if !claims.IsValidAt(time.Now()) {
		return nil, errExpiredToken
	}

	return claims, nil
}
