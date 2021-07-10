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
	s.r = NewRouter("/golauth", s.db)
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
