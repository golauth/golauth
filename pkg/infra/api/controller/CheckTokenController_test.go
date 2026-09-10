package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/application/token/mock"
	"github.com/golauth/golauth/pkg/infra/api/apictx"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type CheckTokenControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl            *gomock.Controller
	validateToken   *mock.MockValidateToken
	app             *fiber.App
	publishedClaims *model.Claims

	ct CheckTokenController
}

func TestCheckTokenControllerSuite(t *testing.T) {
	suite.Run(t, new(CheckTokenControllerSuite))
}

func (s *CheckTokenControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.validateToken = mock.NewMockValidateToken(s.ctrl)
	s.publishedClaims = nil

	s.ct = NewCheckTokenController(s.validateToken)
	s.app = fiber.New()
	s.app.Get("/check_token", s.ct.CheckToken)
	// /me reads claims the security middleware publishes; the stub here stands
	// in for that middleware when s.publishedClaims is set.
	s.app.Get("/me", func(ctx fiber.Ctx) error {
		if s.publishedClaims != nil {
			apictx.SetClaims(ctx, s.publishedClaims)
		}
		return ctx.Next()
	}, s.ct.Me)
}

func (s *CheckTokenControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CheckTokenControllerSuite) TestCheckTokenErrExtractToken() {
	r, _ := http.NewRequest("GET", "/check_token", nil)
	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
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
	s.validateToken.EXPECT().Execute(tk).Return(nil, fmt.Errorf("parsed token invalid")).Times(1)

	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
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
	s.validateToken.EXPECT().Execute(tk).Return(fullClaims(), nil).Times(1)

	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	var got map[string]any
	s.NoError(json.NewDecoder(resp.Body).Decode(&got))
	s.Equal("admin", got["username"])
	s.Equal("Ada", got["firstName"])
	s.Equal("Lovelace", got["lastName"])
	s.Equal([]any{"ADMIN", "USER"}, got["authorities"])
	s.Equal("8c61f220-8bb8-48b9-b225-d54dfa6503db", got["sub"])
	s.Equal(float64(1893456000), got["exp"])
}

func fullClaims() *model.Claims {
	c := &model.Claims{
		Username:    "admin",
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Authorities: []string{"ADMIN", "USER"},
	}
	c.Subject = "8c61f220-8bb8-48b9-b225-d54dfa6503db"
	c.ExpiresAt = jwt.NewNumericDate(time.Unix(1893456000, 0))
	return c
}

// An invalid/expired token is 401 with the validator's message and no claim
// fields in the body.
func (s *CheckTokenControllerSuite) TestCheckTokenExpiredLeaksNoClaims() {
	tk := "eyJhbGciOiJSUzI1NiJ9.e30.sig"
	r, _ := http.NewRequest("GET", "/check_token", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk))
	s.validateToken.EXPECT().Execute(tk).Return(nil, fmt.Errorf("expired token")).Times(1)

	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.NoError(err)
	s.Equal(http.StatusUnauthorized, resp.StatusCode)

	b, _ := io.ReadAll(resp.Body)
	s.Equal("expired token", string(b))
	s.NotContains(string(b), "username")
	s.NotContains(string(b), "sub")
}

func (s *CheckTokenControllerSuite) TestMeReturnsPublishedClaims() {
	s.publishedClaims = fullClaims()

	r, _ := http.NewRequest("GET", "/me", nil)
	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	var got map[string]any
	s.NoError(json.NewDecoder(resp.Body).Decode(&got))
	s.Equal("admin", got["username"])
	s.Equal("8c61f220-8bb8-48b9-b225-d54dfa6503db", got["sub"])
}

// Me never parses a token itself: with nothing published it is 401.
func (s *CheckTokenControllerSuite) TestMeUnauthenticatedWhenNoClaims() {
	r, _ := http.NewRequest("GET", "/me", nil)
	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.NoError(err)
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}
