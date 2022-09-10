package controller

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/src/application/token"
	"github.com/golauth/golauth/src/application/token/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type CheckTokenControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl          *gomock.Controller
	validateToken *mock.MockValidateToken

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
}

func (s *CheckTokenControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CheckTokenControllerSuite) TestCheckTokenErrExtractToken() {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/check_token", nil)

	s.ct.CheckToken(w, r)
	s.Equal(http.StatusBadRequest, w.Code)
	s.ErrorAs(errors.New(w.Body.String()), &token.ErrBearerTokenExtract)
}

func (s *CheckTokenControllerSuite) TestCheckTokenInvalidToken() {
	tk := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/check_token", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk))
	s.validateToken.EXPECT().Execute(tk).Return(fmt.Errorf("parsed token invalid")).Times(1)

	s.ct.CheckToken(w, r)
	s.Equal(http.StatusUnauthorized, w.Code)
	expected := errors.New("parsed token invalid")
	s.ErrorAs(errors.New(w.Body.String()), &expected)
}

func (s *CheckTokenControllerSuite) TestCheckTokenOk() {
	tk := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/check_token", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk))
	s.validateToken.EXPECT().Execute(tk).Return(nil).Times(1)

	s.ct.CheckToken(w, r)
	s.Equal(http.StatusOK, w.Code)
}
