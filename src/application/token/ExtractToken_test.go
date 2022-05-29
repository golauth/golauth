package token

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strings"
	"testing"
)

type ExtractTokenSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
}

func TestExtractToken(t *testing.T) {
	suite.Run(t, new(ExtractTokenSuite))
}

func (s *ExtractTokenSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
}

func (s *ExtractTokenSuite) TearDownTest() {
	s.mockCtrl.Finish()
}
func (s ExtractTokenSuite) TestExtractTokenOk() {
	tk := "aeyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"
	reader := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPost, "/check_token", reader)
	s.NoError(err)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tk))
	extracted, err := ExtractToken(req)
	s.NoError(err)
	s.NotEmpty(extracted)
	s.Equal(tk, extracted)
}

func (s ExtractTokenSuite) TestExtractTokenNotOk() {
	reader := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPost, "/check_token", reader)
	if err != nil {
		s.T().Fatal("error when creating request: %w", err)
	}
	extracted, err := ExtractToken(req)
	s.Error(err)
	s.Empty(extracted)
	s.ErrorAs(err, &ErrBearerTokenExtract)
}
