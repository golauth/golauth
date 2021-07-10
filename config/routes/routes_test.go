package routes

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http/httptest"
	"testing"
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

type RoutesSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	db       *sql.DB
	mockDB   sqlmock.Sqlmock
	r        Router
	router   *mux.Router
	recorder *httptest.ResponseRecorder
}

func TestRoutes(t *testing.T) {
	suite.Run(t, new(RoutesSuite))
}

func (s *RoutesSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	var err error
	s.db, s.mockDB, err = sqlmock.New()
	s.NoError(err)
	s.r = NewRouter("/golauth", s.db, privBytes, pubBytes)
	s.router = mux.NewRouter()
	s.recorder = httptest.NewRecorder()
}

func (s *RoutesSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *RoutesSuite) TestRouteSignupRegistered() {
	s.r.RegisterRoutes(s.router)
	r := s.router.GetRoute("signup")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "POST", "OPTIONS")
	s.NoError(err)
	s.Equal("/signup", tpl)
}

func (s *RoutesSuite) TestRouteTokenRegistered() {
	s.r.RegisterRoutes(s.router)
	r := s.router.GetRoute("token")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "POST", "OPTIONS")
	s.NoError(err)
	s.Equal("/token", tpl)
}

func (s *RoutesSuite) TestRouteCheckTokenRegistered() {
	s.r.RegisterRoutes(s.router)
	r := s.router.GetRoute("checkToken")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "GET", "OPTIONS")
	s.NoError(err)
	s.Equal("/check_token", tpl)
}

func (s *RoutesSuite) TestRouteGetUserRegistered() {
	s.r.RegisterRoutes(s.router)
	r := s.router.GetRoute("getUser")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "GET", "OPTIONS")
	s.NoError(err)
	s.Equal("/users/{username}", tpl)
}

func (s *RoutesSuite) TestRouteAddRoleToUserRegistered() {
	s.r.RegisterRoutes(s.router)
	r := s.router.GetRoute("addRoleToUser")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "POST", "OPTIONS")
	s.NoError(err)
	s.Equal("/users/{username}/add-role", tpl)
}

func (s *RoutesSuite) TestRouteAddRoleRegistered() {
	s.r.RegisterRoutes(s.router)
	r := s.router.GetRoute("addRole")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "POST", "OPTIONS")
	s.NoError(err)
	s.Equal("/roles", tpl)
}

func (s *RoutesSuite) TestRouteEditRoleRegistered() {
	s.r.RegisterRoutes(s.router)
	r := s.router.GetRoute("editRole")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "PUT", "OPTIONS")
	s.NoError(err)
	s.Equal("/roles/{id}", tpl)
}

func (s *RoutesSuite) validateMethods(r *mux.Route, args ...string) {
	for _, m := range args {
		if !s.isRegistered(r, m) {
			s.Failf("method not found", "Method %s not registered for route %s", m, r.GetName())
		}
	}
}

func (s *RoutesSuite) isRegistered(r *mux.Route, method string) bool {
	routeMethods, err := r.GetMethods()
	s.NoError(err)
	for _, rm := range routeMethods {
		if rm == method {
			return true
		}
	}
	return false
}
