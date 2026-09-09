package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/domain/entity"
	factorymock "github.com/golauth/golauth/pkg/domain/factory/mock"
	repomock "github.com/golauth/golauth/pkg/domain/repository/mock"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// adminPassword is the plaintext behind adminPasswordHash, matching the seed
// used by the initial_data migration.
const (
	adminPassword     = "admin123"
	adminPasswordHash = "$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22"
)

// publicPaths are the only paths the API is meant to serve without a token.
var publicPaths = map[string]bool{
	"/auth/token":       true,
	"/auth/check_token": true,
	"/auth/signup":      true,
}

// RoutesSuite exercises the router exactly as main.go builds it. The
// vulnerability fixed here (GHSA-p34g-m47x-q2m4) was invisible to the
// middleware's own unit test because that test assembles its own app with the
// middleware registered before the route; only the real Config() reproduced
// the faulty ordering.
type RoutesSuite struct {
	suite.Suite
	*require.Assertions
	ctrl *gomock.Controller

	userRepository          *repomock.MockUserRepository
	roleRepository          *repomock.MockRoleRepository
	userRoleRepository      *repomock.MockUserRoleRepository
	userAuthorityRepository *repomock.MockUserAuthorityRepository

	app    *fiber.App
	userID uuid.UUID
}

func TestRoutesSuite(t *testing.T) {
	suite.Run(t, new(RoutesSuite))
}

func (s *RoutesSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.userRepository = repomock.NewMockUserRepository(s.ctrl)
	s.roleRepository = repomock.NewMockRoleRepository(s.ctrl)
	s.userRoleRepository = repomock.NewMockUserRoleRepository(s.ctrl)
	s.userAuthorityRepository = repomock.NewMockUserAuthorityRepository(s.ctrl)

	repoFactory := factorymock.NewMockRepositoryFactory(s.ctrl)
	repoFactory.EXPECT().NewUserRepository().Return(s.userRepository).AnyTimes()
	repoFactory.EXPECT().NewRoleRepository().Return(s.roleRepository).AnyTimes()
	repoFactory.EXPECT().NewUserRoleRepository().Return(s.userRoleRepository).AnyTimes()
	repoFactory.EXPECT().NewUserAuthorityRepository().Return(s.userAuthorityRepository).AnyTimes()

	s.userID = uuid.New()
	s.app = NewRouter(repoFactory).Config()
}

func (s *RoutesSuite) TearDownTest() {
	s.ctrl.Finish()
}

// login drives the real public token endpoint, so the tokens under test are
// produced by the same code path an attacker would use.
func (s *RoutesSuite) login(authorities ...string) string {
	s.userRepository.EXPECT().
		FindByUsername(gomock.Any(), "admin").
		Return(&entity.User{ID: s.userID, Username: "admin", Password: adminPasswordHash}, nil).
		Times(1)
	s.userAuthorityRepository.EXPECT().
		FindAuthoritiesByUserID(gomock.Any(), s.userID).
		Return(authorities, nil).
		Times(1)

	body := fmt.Sprintf(`{"username":"admin","password":%q}`, adminPassword)
	req, _ := http.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	s.Equal(http.StatusOK, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	s.NoError(err)
	var tr model.TokenResponse
	s.NoError(json.Unmarshal(raw, &tr))
	s.NotEmpty(tr.AccessToken)
	return tr.AccessToken
}

// do issues a request and reports the status code. It owns the response body
// so no caller has to remember to close it.
func (s *RoutesSuite) do(method, path, token string) int {
	req, _ := http.NewRequest(method, path, strings.NewReader("{}"))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	if token != "" {
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	}
	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

// concretePath substitutes route parameters so a declared route can actually
// be requested.
func concretePath(route string, userID uuid.UUID) string {
	r := strings.ReplaceAll(route, ":id", userID.String())
	return strings.ReplaceAll(r, ":name", "ADMIN")
}

// TestEveryNonPublicRouteRequiresAuthentication is the regression barrier for
// GHSA-p34g-m47x-q2m4. It is driven by the router's own route table rather
// than a hand-written list, so a route added later is covered automatically
// and cannot silently ship unauthenticated.
func (s *RoutesSuite) TestEveryNonPublicRouteRequiresAuthentication() {
	checked := 0
	for _, route := range s.app.GetRoutes(true) {
		if route.Method == http.MethodHead || route.Method == http.MethodOptions {
			continue
		}
		if !strings.HasPrefix(route.Path, pathPrefix) || publicPaths[route.Path] {
			continue
		}

		s.Equal(http.StatusUnauthorized, s.do(route.Method, concretePath(route.Path, s.userID), ""),
			"%s %s must reject an unauthenticated request", route.Method, route.Path)
		checked++
	}
	s.Greater(checked, 0, "no protected route was exercised; the guard would be vacuous")
}

// TestPublicRoutesRemainReachable guards the other direction: fixing the path
// comparison in isPrivateURI is what keeps login working once the middleware
// actually runs.
func (s *RoutesSuite) TestPublicRoutesRemainReachable() {
	s.Run("signup", func() {
		newUser := &entity.User{ID: uuid.New(), Username: "someone"}
		roleID := uuid.New()
		s.userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(newUser, nil).Times(1)
		s.roleRepository.EXPECT().FindByName(gomock.Any(), "USER").
			Return(&entity.Role{ID: roleID, Name: "USER"}, nil).Times(1)
		s.userRoleRepository.EXPECT().AddUserRole(gomock.Any(), newUser.ID, roleID).Return(nil).Times(1)

		body := `{"username":"someone","firstName":"Some","lastName":"One","email":"some@one.test","document":"1","password":"pass123456"}`
		req, _ := http.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusCreated, resp.StatusCode)
	})

	s.Run("token", func() {
		s.NotEmpty(s.login("USER")) // login asserts 200 on the public token route
	})

	s.Run("check_token", func() {
		// Reached its own handler, which reports the missing header as 400.
		// A 401 here would mean the security middleware intercepted a public
		// route -- the failure mode of a half-applied fix.
		s.Equal(http.StatusBadRequest, s.do(http.MethodGet, "/auth/check_token", ""))
	})
}

// TestAdvisoryExploitChain replays the published PoC. Authentication alone
// does not stop it: a self-registered account holds a valid token, so the
// escalation steps have to fail on authority, not just on the missing header.
func (s *RoutesSuite) TestAdvisoryExploitChain() {
	s.Run("unauthenticated", func() {
		s.Equal(http.StatusUnauthorized, s.do(http.MethodGet, "/auth/roles/ADMIN", ""))
		s.Equal(http.StatusUnauthorized,
			s.do(http.MethodPost, "/auth/users/"+s.userID.String()+"/add-role", ""))
	})

	s.Run("authenticated as plain USER", func() {
		userToken := s.login("USER")
		s.Equal(http.StatusForbidden, s.do(http.MethodGet, "/auth/roles/ADMIN", userToken),
			"ADMIN role lookup must be denied to a non-admin token")
		s.Equal(http.StatusForbidden,
			s.do(http.MethodPost, "/auth/users/"+s.userID.String()+"/add-role", userToken),
			"self role assignment must be denied to a non-admin token")
	})
}

func (s *RoutesSuite) TestAdminMayManageRoles() {
	adminToken := s.login("ADMIN", "USER")
	roleID := uuid.New()
	s.roleRepository.EXPECT().FindByName(gomock.Any(), "ADMIN").
		Return(&entity.Role{ID: roleID, Name: "ADMIN"}, nil).Times(1)

	s.Equal(http.StatusOK, s.do(http.MethodGet, "/auth/roles/ADMIN", adminToken))
}

func (s *RoutesSuite) TestFindUserByIdIsSelfOrAdmin() {
	s.Run("owner reaches its own user", func() {
		userToken := s.login("USER")
		s.userRepository.EXPECT().FindByID(gomock.Any(), s.userID).
			Return(&entity.User{ID: s.userID, Username: "admin"}, nil).Times(1)

		s.Equal(http.StatusOK, s.do(http.MethodGet, "/auth/users/"+s.userID.String(), userToken))
	})

	s.Run("non admin cannot reach another user", func() {
		userToken := s.login("USER")
		s.Equal(http.StatusForbidden,
			s.do(http.MethodGet, "/auth/users/"+uuid.New().String(), userToken))
	})

	s.Run("admin reaches any user", func() {
		adminToken := s.login("ADMIN", "USER")
		other := uuid.New()
		s.userRepository.EXPECT().FindByID(gomock.Any(), other).
			Return(&entity.User{ID: other, Username: "other"}, nil).Times(1)

		s.Equal(http.StatusOK, s.do(http.MethodGet, "/auth/users/"+other.String(), adminToken))
	})
}

// TestPreflightIsNotAuthenticated keeps cors registered ahead of the security
// middleware: a browser preflight carries no Authorization header.
func (s *RoutesSuite) TestPreflightIsNotAuthenticated() {
	req, _ := http.NewRequest(http.MethodOptions, "/auth/roles/ADMIN", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)

	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	s.NotEqual(http.StatusUnauthorized, resp.StatusCode)
}
