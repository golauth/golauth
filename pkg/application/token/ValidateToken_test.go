package token

import (
	"fmt"
	"github.com/golauth/golauth/pkg/domain/entity"
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
	key := GeneratePrivateKey()
	s.jwtToken = NewGenerateJwtToken(key)
	s.validateToken = NewValidateToken(key)

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
	TokenExpirationTime = 30
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
	TokenExpirationTime = -1
	expiredToken, err := s.jwtToken.Execute(s.user, []string{"ADMIN"})
	s.NoError(err)
	claims, err := s.validateToken.Execute(expiredToken)
	s.Error(err)
	s.Nil(claims)
	s.ErrorIs(err, errExpiredToken)
}
