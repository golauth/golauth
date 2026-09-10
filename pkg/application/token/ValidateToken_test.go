package token

import (
	"fmt"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/infra/keys"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type ValidateTokenSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	ks            *keys.KeySet
	jwtToken      GenerateJwtToken
	validateToken ValidateToken
	user          *entity.User
}

func TestValidateToken(t *testing.T) {
	suite.Run(t, new(ValidateTokenSuite))
}

func (s *ValidateTokenSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ks = keys.Generate()
	s.jwtToken = NewGenerateJwtToken(s.ks.Current, 30*time.Minute)
	s.validateToken = NewValidateToken(s.ks)

	s.user = &entity.User{
		ID:           uuid.New(),
		Username:     "user",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@il.com",
		Document:     "1234",
		Password:     "1234",
		Enabled:      true,
		CreationDate: time.Now(),
	}
}

func (s *ValidateTokenSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ValidateTokenSuite) TestValidateTokenOk() {
	token, err := s.jwtToken.Execute(s.user, []string{"ADMIN"})
	s.NoError(err)
	claims, err := s.validateToken.Execute(fmt.Sprintf("%v", token))
	s.NoError(err)
	s.Equal(s.user.Username, claims.Username)
	s.Equal(s.user.ID.String(), claims.Subject)
	s.Equal([]string{"ADMIN"}, claims.Authorities)
}

func (s *ValidateTokenSuite) TestValidateTokenInvalidFormat() {
	claims, err := s.validateToken.Execute("invalidTokenFormat")
	s.Error(err)
	s.Nil(claims)
	s.EqualError(err, "could not parse and verify strToken: jwt: token format is not valid")
}

func (s *ValidateTokenSuite) TestValidateTokenErrExpiredToken() {
	// Sign with the suite's current key (so verification reaches the expiry
	// check) but with a TTL that has already lapsed by the time we validate.
	expired := NewGenerateJwtToken(s.ks.Current, time.Nanosecond)
	time.Sleep(time.Millisecond)
	expiredToken, err := expired.Execute(s.user, []string{"ADMIN"})
	s.NoError(err)
	claims, err := s.validateToken.Execute(expiredToken)
	s.Error(err)
	s.Nil(claims)
	s.ErrorIs(err, errExpiredToken)
}

// TestValidateTokenSignedWithPreviousKey is the rotation barrier: a token
// signed by a key that has been retired to the verify-only slot still
// validates, while the current key signs new tokens.
func (s *ValidateTokenSuite) TestValidateTokenSignedWithPreviousKey() {
	previous := keys.Generate().Current
	current := keys.Generate().Current
	ks := &keys.KeySet{Current: current, Previous: []*keys.SigningKey{previous}}

	oldToken, err := NewGenerateJwtToken(previous, 30*time.Minute).Execute(s.user, []string{"ADMIN"})
	s.NoError(err)
	claims, err := NewValidateToken(ks).Execute(oldToken)
	s.NoError(err)
	s.Equal(s.user.Username, claims.Username)

	newToken, err := NewGenerateJwtToken(current, 30*time.Minute).Execute(s.user, []string{"ADMIN"})
	s.NoError(err)
	_, err = NewValidateToken(ks).Execute(newToken)
	s.NoError(err)
}

// TestValidateTokenUnknownKeyID rejects a token whose kid names no key we hold.
func (s *ValidateTokenSuite) TestValidateTokenUnknownKeyID() {
	strayToken, err := NewGenerateJwtToken(keys.Generate().Current, 30*time.Minute).Execute(s.user, []string{"ADMIN"})
	s.NoError(err)

	claims, err := s.validateToken.Execute(strayToken)
	s.Nil(claims)
	s.ErrorIs(err, errUnknownKeyID)
}
