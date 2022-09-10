package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/src/application/token/mock"
	"github.com/golauth/golauth/src/domain/entity"
	repoMock "github.com/golauth/golauth/src/domain/repository/mock"
	"github.com/golauth/golauth/src/infra/api/controller/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TokenControllerSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	ctx           context.Context
	uRepo         *repoMock.MockUserRepository
	uaRepo        *repoMock.MockUserAuthorityRepository
	generateToken *mock.MockGenerateToken

	ctrl TokenController
}

func TestTokenController(t *testing.T) {
	suite.Run(t, new(TokenControllerSuite))
}

func (s *TokenControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.uRepo = repoMock.NewMockUserRepository(s.mockCtrl)
	s.uaRepo = repoMock.NewMockUserAuthorityRepository(s.mockCtrl)
	s.generateToken = mock.NewMockGenerateToken(s.mockCtrl)

	s.ctrl = NewTokenController(s.uRepo, s.uaRepo, s.generateToken)
}

func (s *TokenControllerSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *TokenControllerSuite) TestTokenFormOk() {
	username := "admin"
	password := "123456"
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", strings.NewReader(fmt.Sprintf("username=%s&password=%s", username, password)))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s.generateToken.EXPECT().Execute(r.Context(), username, password).Return(&entity.Token{AccessToken: token}, nil).Times(1)

	s.ctrl.Token(w, r)
	s.Equal(http.StatusOK, w.Code)

	var result model.TokenResponse
	_ = json.NewDecoder(w.Body).Decode(&result)
	s.Equal(token, result.AccessToken)
}

func (s *TokenControllerSuite) TestTokenJsonOk() {
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

	s.generateToken.EXPECT().Execute(r.Context(), username, password).Return(&entity.Token{AccessToken: token}, nil).Times(1)

	s.ctrl.Token(w, r)
	s.Equal(http.StatusOK, w.Code)

	var result model.TokenResponse
	_ = json.NewDecoder(w.Body).Decode(&result)
	s.Equal(token, result.AccessToken)
}

func (s *TokenControllerSuite) TestTokenJsonNotOk() {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", strings.NewReader("{foo:bar}"))
	r.Header.Set("Content-Type", "application/json")

	s.ctrl.Token(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "json decoder error")
}

func (s *TokenControllerSuite) TestTokenMethodNotAllowed() {
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

func (s *TokenControllerSuite) TestTokenErrParseForm() {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s.ctrl.Token(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "parse form error: missing form body")
}

func (s *TokenControllerSuite) TestTokenErrGenerateToken() {
	username := "admin"
	password := "123456"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/token", strings.NewReader(fmt.Sprintf("username=%s&password=%s", username, password)))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s.generateToken.EXPECT().Execute(s.ctx, username, password).Return(nil, fmt.Errorf("could not find user by username admin")).Times(1)

	s.ctrl.Token(w, r)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "could not find user by username admin")
}
