package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	repoMock "golauth/infrastructure/repository/mock"
	"golauth/model"
	svcMock "golauth/usecase/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TokenControllerSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	uRepo  *repoMock.MockUserRepository
	uaRepo *repoMock.MockUserAuthorityRepository
	tkSvc  *svcMock.MockTokenService
	uSvc   *svcMock.MockUserService

	ctrl TokenController
}

func TestTokenController(t *testing.T) {
	suite.Run(t, new(TokenControllerSuite))
}

func (s *TokenControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.uRepo = repoMock.NewMockUserRepository(s.mockCtrl)
	s.uaRepo = repoMock.NewMockUserAuthorityRepository(s.mockCtrl)
	s.tkSvc = svcMock.NewMockTokenService(s.mockCtrl)
	s.uSvc = svcMock.NewMockUserService(s.mockCtrl)

	s.ctrl = NewTokenController(s.uRepo, s.uaRepo, s.tkSvc, s.uSvc)
}

func (s *TokenControllerSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s TokenControllerSuite) TestTokenFormOk() {
	username := "admin"
	password := "123456"
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", strings.NewReader(fmt.Sprintf("username=%s&password=%s", username, password)))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s.uSvc.EXPECT().GenerateToken(username, password).Return(model.TokenResponse{AccessToken: token}, nil).Times(1)

	s.ctrl.Token(w, r)
	s.Equal(http.StatusOK, w.Code)

	var result model.TokenResponse
	_ = json.NewDecoder(w.Body).Decode(&result)
	s.Equal(token, result.AccessToken)
}

func (s TokenControllerSuite) TestTokenJsonOk() {
	username := "admin"
	password := "123456"
	login := model.UserLoginRequest{
		Username: username,
		Password: password,
	}
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"

	body, _ := json.Marshal(login)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.uSvc.EXPECT().GenerateToken(username, password).Return(model.TokenResponse{AccessToken: token}, nil).Times(1)

	s.ctrl.Token(w, r)
	s.Equal(http.StatusOK, w.Code)

	var result model.TokenResponse
	_ = json.NewDecoder(w.Body).Decode(&result)
	s.Equal(token, result.AccessToken)
}

func (s TokenControllerSuite) TestTokenJsonNotOk() {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", strings.NewReader("{foo:bar}"))
	r.Header.Set("Content-Type", "application/json")

	s.ctrl.Token(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "json decoder error")
}

func (s TokenControllerSuite) TestTokenMethodNotAllowed() {
	username := "admin"
	password := "123456"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", strings.NewReader(fmt.Sprintf("username=%s&password=%s", username, password)))
	r.Header.Set("Content-Type", "text/html")

	s.ctrl.Token(w, r)
	s.Equal(http.StatusMethodNotAllowed, w.Code)
	errResult := errors.New(w.Body.String())
	s.ErrorAs(ErrContentTypeNotSupported, &errResult)
}

func (s TokenControllerSuite) TestTokenErrParseForm() {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s.ctrl.Token(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "parse form error: missing form body")
}

func (s TokenControllerSuite) TestTokenErrGenerateToken() {
	username := "admin"
	password := "123456"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", strings.NewReader(fmt.Sprintf("username=%s&password=%s", username, password)))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s.uSvc.EXPECT().GenerateToken(username, password).Return(model.TokenResponse{}, fmt.Errorf("could not find user by username admin")).Times(1)

	s.ctrl.Token(w, r)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "could not find user by username admin")
}
