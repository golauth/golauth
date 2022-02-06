package api

import (
	"github.com/golang/mock/gomock"
	factoryMock "github.com/golauth/golauth/domain/factory/mock"
	repoMock "github.com/golauth/golauth/domain/repository/mock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http/httptest"
	"testing"
)

type RoutesSuite struct {
	suite.Suite
	*require.Assertions
	ctrl        *gomock.Controller
	repoFactory *factoryMock.MockRepositoryFactory
	r           Router
	router      *mux.Router
	recorder    *httptest.ResponseRecorder
}

func TestRoutes(t *testing.T) {
	suite.Run(t, new(RoutesSuite))
}

func (s *RoutesSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	var err error
	s.repoFactory = factoryMock.NewMockRepositoryFactory(s.ctrl)
	s.repoFactory.EXPECT().NewUserRepository().AnyTimes().Return(repoMock.NewMockUserRepository(s.ctrl))
	s.repoFactory.EXPECT().NewRoleRepository().AnyTimes().Return(repoMock.NewMockRoleRepository(s.ctrl))
	s.repoFactory.EXPECT().NewUserRoleRepository().AnyTimes().Return(repoMock.NewMockUserRoleRepository(s.ctrl))
	s.repoFactory.EXPECT().NewUserAuthorityRepository().AnyTimes().Return(repoMock.NewMockUserAuthorityRepository(s.ctrl))
	s.NoError(err)
	s.r = NewRouter(s.repoFactory)
	s.router = s.r.Config()
	s.recorder = httptest.NewRecorder()
}

func (s *RoutesSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RoutesSuite) TestRouteSignupRegistered() {
	r := s.router.GetRoute("signup")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "POST", "OPTIONS")
	s.NoError(err)
	s.Equal("/auth/signup", tpl)
}

func (s *RoutesSuite) TestRouteTokenRegistered() {
	r := s.router.GetRoute("token")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "POST", "OPTIONS")
	s.NoError(err)
	s.Equal("/auth/token", tpl)
}

func (s *RoutesSuite) TestRouteCheckTokenRegistered() {
	r := s.router.GetRoute("checkToken")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "GET", "OPTIONS")
	s.NoError(err)
	s.Equal("/auth/check_token", tpl)
}

func (s *RoutesSuite) TestRouteGetUserRegistered() {
	r := s.router.GetRoute("getUser")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "GET", "OPTIONS")
	s.NoError(err)
	s.Equal("/auth/users/{id}", tpl)
}

func (s *RoutesSuite) TestRouteAddRoleToUserRegistered() {
	r := s.router.GetRoute("addRoleToUser")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "POST", "OPTIONS")
	s.NoError(err)
	s.Equal("/auth/users/{id}/add-role", tpl)
}

func (s *RoutesSuite) TestRouteAddRoleRegistered() {
	r := s.router.GetRoute("addRole")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "POST", "OPTIONS")
	s.NoError(err)
	s.Equal("/auth/roles", tpl)
}

func (s *RoutesSuite) TestRouteFindRoleByNameRegistered() {
	r := s.router.GetRoute("findRoleByName")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "GET", "OPTIONS")
	s.NoError(err)
	s.Equal("/auth/roles/{name}", tpl)
}

func (s *RoutesSuite) TestRouteEditRoleRegistered() {
	r := s.router.GetRoute("editRole")
	s.NotNil(r)
	tpl, err := r.GetPathTemplate()
	s.validateMethods(r, "PUT", "OPTIONS")
	s.NoError(err)
	s.Equal("/auth/roles/{id}", tpl)
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
