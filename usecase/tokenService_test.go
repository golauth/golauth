package usecase

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/model"
	"net/http"
	"strings"
	"testing"
	"time"
)

var privBytes = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEAxb5Ug8Y+qQ7LhFqdI9OsBJxlnscV72gCeiy6W2VthNcSde4G
2bt+8IdMUZYd4K+b+JcQ3h4mDzRG8WulzuWZBZ0/6Qjb2muoFg2WIa215Uj3OseJ
gINaOGS9qVb0HyPWxA/CPpUvlatiBNj9UD6mOEuS5gc7xI1mC1AUhb8gZlOwVmT8
RLNF2nmpRysE8OWi/dKeHokV9CkXGuhaPetyiPZktLZVWQp0gWUMSeZ90aeiYHrN
MerY6iAtvNPTpcJTUaTIKj6RehFJWWyi5szSC3lphcWNqIkK/9zoAfe5ac+BxjSH
2PH69XZsPFg3UKe0bylnnTZP3IIL2iavXEW/AQIDAQABAoIBAQCMgjm7iNptdj3W
xixykK3ieN8ce4pymw1nkvC4kNHJWqmbco8bl8cTUpBASNLiHOZPNciei/2vQA5I
7Zzb7vlUq/AFvm26PlUplm3fcHeXfMlv0uk5kBxDhhHeihLdLbIljq0PmyI8z5LO
rwEQS/QAfHLdULZ/a5ne4AA1KSH0krFmH3ffHCceER6tT7RtIZUaWpw4uybubvW3
rcA/RPKTonOuFkGWPD/u4a/Oh1GN3bCtaMUHpU14ydglBzCMX/jS3wWTwMl8sm9E
MKdOIrHw1xSKmh6wC605CiIC+Svf6mNWSm3mfJeCAGRfHw30UlLoovN0ZMgC4jLa
Vnns68AhAoGBAPooKcJg6DoZ+0Z1Qpc6LTqn1kuq3i+qnhHke3EIWm3f9djFPrOp
NnCW92j9LP5EeGF2UUSQW/7jjva/c1DoghHpm0ByXsCDXY3B9NCV78806/1orsZq
ImPHg1kVivvrtkkTvPuXGZ5TwGqmTRZ0Lbbh0LP9nrJraSGzOxJHRpIjAoGBAMpc
wYxQT3m6/rvXagY0jelvifD6BqQXMvPCsssk6foaIbXYGp5R24u0adzpHh79YN6a
8trTMnfgLXdK9YYbJIU0twYNxg9W3Ke4ilyA68maNUJpnoiP7KvbHligMTMKqftB
UnNneiulq2qQ4Q6MK34hK7BfYc4KQ1KTlZyfoWKLAoGBAIeJ/V5RTWI1s5zwad0w
a1MtnwGumeYvxqehKXUL9pszzqvd62RC2blVQsZC7v7xsFv2VIAWy5GmUE7HWr7K
y7bS4QihL0+VnbnyDih6JM4bOYY7Ev90gB+Z+UPqVTy78S9VH38d1oafkFD4vCnf
VumRHph3YWYAppzY1LfJoKYLAoGBALyJ1zpnyORduOBCP2IwrNeFODvwdyeDBdHe
4L4sUmLW3fmSspo3IhnzqX5NI+czo4FDVGlUxHyzvSicCk08FLaW+r8FLjc0crlB
UogFBan7pwuNZEtP7O3hZVClT7GCigSyQ6OKEWWBIUhUW5s2NX96YD4fX/yby0Ww
g4A9qhspAoGBAJ+/i610WZV5q0VsjPprVPnIoE3e0Bhm9MJK5GN9WsVkZ8Sv7z6v
I8z+Xa4SkYYIhmjoosdrDqLvd43GIUxgzhLgZHoBDm+9Sgaxuw3IVDzNKNAF8f2T
RWjftkRYQlUU6oAtc4OrWfk5LWC7He7jT7WxpvBYIBhUQwezR/ttNhB+
-----END RSA PRIVATE KEY-----`)

var pubBytes = []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxb5Ug8Y+qQ7LhFqdI9Os
BJxlnscV72gCeiy6W2VthNcSde4G2bt+8IdMUZYd4K+b+JcQ3h4mDzRG8WulzuWZ
BZ0/6Qjb2muoFg2WIa215Uj3OseJgINaOGS9qVb0HyPWxA/CPpUvlatiBNj9UD6m
OEuS5gc7xI1mC1AUhb8gZlOwVmT8RLNF2nmpRysE8OWi/dKeHokV9CkXGuhaPety
iPZktLZVWQp0gWUMSeZ90aeiYHrNMerY6iAtvNPTpcJTUaTIKj6RehFJWWyi5szS
C3lphcWNqIkK/9zoAfe5ac+BxjSH2PH69XZsPFg3UKe0bylnnTZP3IIL2iavXEW/
AQIDAQAB
-----END PUBLIC KEY-----`)

type TokenServiceSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	svc TokenService
}

func TestTokenService(t *testing.T) {
	suite.Run(t, new(TokenServiceSuite))
}

func (s *TokenServiceSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.svc = NewTokenService(privBytes, pubBytes)
}

func (s *TokenServiceSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s TokenServiceSuite) TestExtractTokenOk() {
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"
	reader := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPost, "/check_token", reader)
	if err != nil {
		s.T().Fatal("error when creating request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	extracted, err := s.svc.ExtractToken(req)
	s.NoError(err)
	s.NotEmpty(extracted)
	s.Equal(token, extracted)
}

func (s TokenServiceSuite) TestExtractTokenNotOk() {
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

func (s TokenServiceSuite) TestValidateTokenOk() {
	user := model.User{
		ID:           0,
		Username:     "user",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@il.com",
		Document:     "1234",
		Password:     "1234",
		Enabled:      true,
		CreationDate: time.Time{},
	}
	token, err := s.svc.GenerateJwtToken(user, []string{"ADMIN"})
	s.NoError(err)
	err = s.svc.ValidateToken(fmt.Sprintf("%v", token))
	s.NoError(err)
}

func (s TokenServiceSuite) TestValidateTokenErrParseClaims() {
	err := s.svc.ValidateToken("nil")
	s.Error(err)
	s.EqualError(err, "error when parse with claims: token contains an invalid number of segments")
}

// ===================================================================================

type TokenServiceInvalidKeysSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
}

func TestTokenServiceInvalidKeys(t *testing.T) {
	suite.Run(t, new(TokenServiceInvalidKeysSuite))
}

func (s *TokenServiceInvalidKeysSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
}

func (s *TokenServiceInvalidKeysSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *TokenServiceInvalidKeysSuite) TestNewTokenServicePrivKeyInvalid() {
	errChan := make(chan error, 1)
	expected := errors.New("could not parse RSA private key from pem: Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key")
	getPanic(errChan, NewTokenService, nil, nil)
	err := <-errChan
	s.Error(err)
	s.Equal(expected.Error(), err.Error())
}

func (s *TokenServiceInvalidKeysSuite) TestNewTokenServicePubKeyInvalid() {
	errChan := make(chan error, 1)
	expected := errors.New("could not parse RSA public key from pem: Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key")
	getPanic(errChan, NewTokenService, privBytes, nil)
	err := <-errChan
	s.Error(err)
	s.Equal(expected.Error(), err.Error())
}

type funcNewTokenService func(privBytes []byte, pubBytes []byte) TokenService

func getPanic(errChan chan error, fn funcNewTokenService, args ...[]byte) {
	defer func() {
		if r := recover(); r == nil {
			errChan <- nil
		} else {
			errChan <- r.(error)
		}
	}()
	fn(args[0], args[1])
}
