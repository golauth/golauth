package token

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strings"
	"testing"
	"time"
)

type ServiceSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	svc UseCase

	user entity.User
}

func TestTokenService(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	key := GeneratePrivateKey()
	s.svc = NewService(key)

	s.user = entity.User{
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

func (s *ServiceSuite) TearDownTest() {
	tokenExpirationTime = 30
	s.mockCtrl.Finish()
}

func (s ServiceSuite) TestExtractTokenOk() {
	token := "aeyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"
	reader := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPost, "/check_token", reader)
	s.NoError(err)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	extracted, err := s.svc.ExtractToken(req)
	s.NoError(err)
	s.NotEmpty(extracted)
	s.Equal(token, extracted)
}

func (s ServiceSuite) TestExtractTokenNotOk() {
	reader := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPost, "/check_token", reader)
	if err != nil {
		s.T().Fatal("error when creating request: %w", err)
	}
	extracted, err := s.svc.ExtractToken(req)
	s.Error(err)
	s.Empty(extracted)
	s.ErrorAs(err, &ErrBearerTokenExtract)
}
