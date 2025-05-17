package controller

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/application/token/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"testing"
)

type CheckTokenControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl          *gomock.Controller
	validateToken *mock.MockValidateToken
	app           *fiber.App

	ct CheckTokenController
}

func TestCheckTokenControllerSuite(t *testing.T) {
	suite.Run(t, new(CheckTokenControllerSuite))
}

func (s *CheckTokenControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.validateToken = mock.NewMockValidateToken(s.ctrl)

	s.ct = NewCheckTokenController(s.validateToken)
	s.app = fiber.New()
	s.app.Get("/check_token", s.ct.CheckToken)
}

func (s *CheckTokenControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CheckTokenControllerSuite) TestCheckTokenErrExtractToken() {
	r, _ := http.NewRequest("GET", "/check_token", nil)
	resp, err := s.app.Test(r, -1)
	s.NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Equal(token.ErrBearerTokenExtract.Error(), string(b))
}

func (s *CheckTokenControllerSuite) TestCheckTokenInvalidToken() {
	tk := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"

	r, _ := http.NewRequest("GET", "/check_token", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk))
	s.validateToken.EXPECT().Execute(tk).Return(fmt.Errorf("parsed token invalid")).Times(1)

	resp, err := s.app.Test(r, -1)
	s.NoError(err)
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
	expectedMsg := "parsed token invalid"
	b, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Equal(expectedMsg, string(b))
}

func (s *CheckTokenControllerSuite) TestCheckTokenOk() {
	tk := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"

	r, _ := http.NewRequest("GET", "/check_token", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk))
	s.validateToken.EXPECT().Execute(tk).Return(nil).Times(1)

	resp, err := s.app.Test(r, -1)
	s.NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)
}
