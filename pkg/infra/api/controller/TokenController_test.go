package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/pkg/application/token/mock"
	"github.com/golauth/golauth/pkg/domain/entity"
	repoMock "github.com/golauth/golauth/pkg/domain/repository/mock"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
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
	app  *fiber.App
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
	s.app = fiber.New()
	s.app.Post("/token", s.ctrl.Token)
}

func (s *TokenControllerSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *TokenControllerSuite) TestTokenFormOk() {
	username := "admin"
	password := "123456"
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"

	r, _ := http.NewRequest("POST", "/token", strings.NewReader(fmt.Sprintf("username=%s&password=%s", username, password)))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s.generateToken.EXPECT().Execute(r.Context(), username, password).Return(&entity.Token{AccessToken: token}, nil).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusOK, resp.StatusCode)

	var result model.TokenResponse
	_ = json.NewDecoder(resp.Body).Decode(&result)
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
	r, _ := http.NewRequest("POST", "/token", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.generateToken.EXPECT().Execute(r.Context(), username, password).Return(&entity.Token{AccessToken: token}, nil).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusOK, resp.StatusCode)

	var result model.TokenResponse
	_ = json.NewDecoder(resp.Body).Decode(&result)
	s.Equal(token, result.AccessToken)
}

func (s *TokenControllerSuite) TestTokenJsonNotOk() {
	r, _ := http.NewRequest("POST", "/token", strings.NewReader("{foo:bar}"))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Contains(string(b), "json decoder error")
}

func (s *TokenControllerSuite) TestTokenMethodNotAllowed() {
	username := "admin"
	password := "123456"

	r, _ := http.NewRequest("POST", "/token", strings.NewReader(fmt.Sprintf("username=%s&password=%s", username, password)))
	r.Header.Set("Content-Type", "text/html")

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Equal(ErrContentTypeNotSupported.Error(), string(b))
}

func (s *TokenControllerSuite) TestTokenErrParseForm() {
	r, _ := http.NewRequest("POST", "/token", nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Contains(string(b), ErrMissingBodyData.Error())
}

func (s *TokenControllerSuite) TestTokenErrGenerateToken() {
	username := "admin"
	password := "123456"

	r, _ := http.NewRequest("POST", "/token", strings.NewReader(fmt.Sprintf("username=%s&password=%s", username, password)))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s.generateToken.EXPECT().Execute(s.ctx, username, password).Return(nil, fmt.Errorf("could not find user by username admin")).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}
