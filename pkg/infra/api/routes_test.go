package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/cristalhq/jwt/v3"
	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/keys"
	"github.com/golauth/golauth/pkg/application/token/claims"
	"github.com/golauth/golauth/pkg/domain/entity"
	factorymock "github.com/golauth/golauth/pkg/domain/factory/mock"
	repomock "github.com/golauth/golauth/pkg/domain/repository/mock"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/golauth/golauth/pkg/infra/logging"
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
	"/auth/token":                 true,
	"/auth/token/refresh":         true,
	"/auth/check_token":           true,
	"/auth/signup":                true,
	"/auth/.well-known/jwks.json": true,
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
	loginAttemptRepository  *repomock.MockLoginAttemptRepository
	refreshTokenRepository  *repomock.MockRefreshTokenRepository

	repoFactory *factorymock.MockRepositoryFactory
	keySet      *keys.KeySet

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
	s.loginAttemptRepository = repomock.NewMockLoginAttemptRepository(s.ctrl)
	s.refreshTokenRepository = repomock.NewMockRefreshTokenRepository(s.ctrl)

	repoFactory := factorymock.NewMockRepositoryFactory(s.ctrl)
	s.repoFactory = repoFactory
	repoFactory.EXPECT().NewUserRepository().Return(s.userRepository).AnyTimes()
	repoFactory.EXPECT().NewRoleRepository().Return(s.roleRepository).AnyTimes()
	repoFactory.EXPECT().NewUserRoleRepository().Return(s.userRoleRepository).AnyTimes()
	repoFactory.EXPECT().NewUserAuthorityRepository().Return(s.userAuthorityRepository).AnyTimes()
	repoFactory.EXPECT().NewLoginAttemptRepository().Return(s.loginAttemptRepository).AnyTimes()
	repoFactory.EXPECT().NewRefreshTokenRepository().Return(s.refreshTokenRepository).AnyTimes()

	s.userID = uuid.New()
	s.keySet = keys.Generate()
	s.app = NewRouter(repoFactory, s.keySet).Config()
}

func (s *RoutesSuite) TearDownTest() {
	s.ctrl.Finish()
}

// passthroughRefreshCreate stands in for the refresh-token repository at login:
// it stamps an id and echoes the row back, so the real GenerateToken use case
// completes without a database.
func passthroughRefreshCreate(_ context.Context, rt *entity.RefreshToken) (*entity.RefreshToken, error) {
	rt.ID = uuid.New()
	return rt, nil
}

// login drives the real public token endpoint, so the tokens under test are
// produced by the same code path an attacker would use.
func (s *RoutesSuite) login(authorities ...string) string {
	s.userRepository.EXPECT().
		FindByUsername(gomock.Any(), "admin").
		Return(&entity.User{ID: s.userID, Username: "admin", Password: adminPasswordHash, Enabled: true}, nil).
		Times(1)
	s.userAuthorityRepository.EXPECT().
		FindAuthoritiesByUserID(gomock.Any(), s.userID).
		Return(authorities, nil).
		Times(1)
	s.loginAttemptRepository.EXPECT().Get(gomock.Any(), s.userID).Return(nil, nil).Times(1)
	s.refreshTokenRepository.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(passthroughRefreshCreate).Times(1)

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

// TestErrorContractOnEveryProtectedRoute drives the router's own route table and
// asserts that the induced failure (a missing token) comes back as the one
// documented error envelope on every route: a JSON body with error.code,
// error.message and a request id that also appears in the X-Request-Id header.
func (s *RoutesSuite) TestErrorContractOnEveryProtectedRoute() {
	checked := 0
	for _, route := range s.app.GetRoutes(true) {
		if route.Method == http.MethodHead || route.Method == http.MethodOptions {
			continue
		}
		if !strings.HasPrefix(route.Path, pathPrefix) || publicPaths[route.Path] {
			continue
		}

		req, _ := http.NewRequest(route.Method, concretePath(route.Path, s.userID), strings.NewReader("{}"))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
		s.Require().NoError(err)

		var body struct {
			Error struct {
				Code, Message, RequestID string
			} `json:"error"`
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		s.Require().NoError(json.Unmarshal(raw, &body), "%s %s: body is not the error envelope: %s", route.Method, route.Path, raw)

		s.Equal("unauthorized", body.Error.Code, "%s %s", route.Method, route.Path)
		s.NotEmpty(body.Error.Message, "%s %s", route.Method, route.Path)
		s.NotEmpty(body.Error.RequestID, "%s %s: no request id in body", route.Method, route.Path)
		s.Equal(resp.Header.Get("X-Request-Id"), body.Error.RequestID, "%s %s: body/header id mismatch", route.Method, route.Path)
		s.NotContains(string(raw), "sql", "%s %s: sql fragment leaked", route.Method, route.Path)
		checked++
	}
	s.Greater(checked, 0)
}

// TestFailedLoginIsAuditedUnderTheRequestID proves the id threads end to end:
// the middleware assigns it, it reaches the GenerateToken use case through the
// request context, and the login_failed audit event carries the same value the
// caller got back in X-Request-Id.
func (s *RoutesSuite) TestFailedLoginIsAuditedUnderTheRequestID() {
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(logging.WithContext(slog.NewJSONHandler(buf, nil))))
	defer slog.SetDefault(prev)

	s.userRepository.EXPECT().FindByUsername(gomock.Any(), "admin").
		Return(nil, fmt.Errorf("no such user")).Times(1)

	body := `{"username":"admin","password":"wrong"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Require().NoError(err)
	_ = resp.Body.Close()
	s.Equal(http.StatusUnauthorized, resp.StatusCode)

	reqID := resp.Header.Get("X-Request-Id")
	s.Require().NotEmpty(reqID)

	var found map[string]any
	for _, raw := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var m map[string]any
		s.Require().NoError(json.Unmarshal([]byte(raw), &m))
		if m["event"] == "login_failed" {
			found = m
		}
	}
	s.Require().NotNil(found, "no login_failed audit event was emitted")
	s.Equal("unknown_user", found["outcome"])
	s.Equal(reqID, found["request_id"], "audit event not tied to the request id")
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

		body := `{"username":"someone","firstName":"Some","lastName":"One","email":"some@one.test","document":"1","password":"pass1234567890"}`
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

	s.Run("token/refresh", func() {
		// Public: an empty body reaches the handler's own 400. A 401 would mean
		// the security middleware intercepted it.
		s.Equal(http.StatusBadRequest, s.do(http.MethodPost, "/auth/token/refresh", ""))
	})
}

// TestLogoutRoutesAreAuthenticated: /auth/logout and /auth/logout/all sit
// behind the security middleware (401 without a token) but, with a valid token,
// reach their own handlers.
func (s *RoutesSuite) TestLogoutRoutesAreAuthenticated() {
	s.Equal(http.StatusUnauthorized, s.do(http.MethodPost, "/auth/logout", ""))
	s.Equal(http.StatusUnauthorized, s.do(http.MethodPost, "/auth/logout/all", ""))

	token := s.login("USER")

	// Empty body -> the logout handler's own 400, proving it ran.
	s.Equal(http.StatusBadRequest, s.do(http.MethodPost, "/auth/logout", token))

	// logout/all revokes by subject; the real use case calls the repo.
	s.refreshTokenRepository.EXPECT().RevokeAllForUser(gomock.Any(), s.userID).Return(int64(1), nil).Times(1)
	s.Equal(http.StatusNoContent, s.do(http.MethodPost, "/auth/logout/all", token))
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

// TestTokenValidatesOnASecondReplica is the regression barrier for the
// multi-replica bug: NewRouter used to mint a fresh in-memory key per process,
// so a token from one instance was rejected by every other. Two routers built
// from the same key set must accept each other's tokens.
func (s *RoutesSuite) TestTokenValidatesOnASecondReplica() {
	replica := NewRouter(s.repoFactory, s.keySet).Config()

	accessToken := s.login("USER") // minted by s.app

	req, _ := http.NewRequest(http.MethodGet, "/auth/check_token", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+accessToken)
	resp, err := replica.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	s.Equal(http.StatusOK, resp.StatusCode)

	var c claims.Claims
	s.NoError(json.NewDecoder(resp.Body).Decode(&c))
	s.Equal(s.userID.String(), c.Subject, "introspection returns the verified subject")
}

// TestMeReturnsTheAuthenticatedIdentity: /auth/me is behind the security
// middleware (401 without a token) and, with one, echoes the claims the
// middleware verified -- no database call.
func (s *RoutesSuite) TestMeReturnsTheAuthenticatedIdentity() {
	s.Equal(http.StatusUnauthorized, s.do(http.MethodGet, "/auth/me", ""))

	accessToken := s.login("USER")
	req, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+accessToken)
	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	s.Equal(http.StatusOK, resp.StatusCode)

	var c claims.Claims
	s.NoError(json.NewDecoder(resp.Body).Decode(&c))
	s.Equal(s.userID.String(), c.Subject)
	s.Equal("admin", c.Username)
	s.Equal([]string{"USER"}, c.Authorities)
}

// TestJWKSIsPublicAndMatchesMintedTokens checks the endpoint is reachable
// without a token and that its kid is the one stamped on a freshly minted
// token, so an offline consumer can pick the right key.
func (s *RoutesSuite) TestJWKSIsPublicAndMatchesMintedTokens() {
	req, _ := http.NewRequest(http.MethodGet, "/auth/.well-known/jwks.json", nil)
	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	s.Equal(http.StatusOK, resp.StatusCode)

	var doc keys.JWKS
	s.NoError(json.NewDecoder(resp.Body).Decode(&doc))
	s.NotEmpty(doc.Keys)

	accessToken := s.login("USER")
	parsed, err := jwt.ParseString(accessToken)
	s.NoError(err)
	s.Equal(s.keySet.Current.KID, parsed.Header().KeyID)

	found := false
	for _, k := range doc.Keys {
		if k.Kid == parsed.Header().KeyID {
			found = true
			s.Equal("RSA", k.Kty)
			s.Equal("sig", k.Use)
			s.Equal("RS512", k.Alg)
			s.NotEmpty(k.N)
			s.Equal("AQAB", k.E)
		}
	}
	s.True(found, "minted token kid %q not present in JWKS", parsed.Header().KeyID)
}

// TestTokenSurvivesPrincipalDeactivation documents a known limitation: token
// validation checks only the signature and expiry, never the current enabled
// state of the user or its roles. A token minted while the account was active
// keeps working until it expires even after the account is disabled in the
// database. Immediate revocation is Plan 05's job.
func (s *RoutesSuite) TestTokenSurvivesPrincipalDeactivation() {
	adminToken := s.login("ADMIN", "USER")

	// No repository call re-checks enabled here; the middleware only verifies
	// the token. A second admin action with the same token still authorizes.
	roleID := uuid.New()
	s.roleRepository.EXPECT().FindByName(gomock.Any(), "ADMIN").
		Return(&entity.Role{ID: roleID, Name: "ADMIN"}, nil).Times(1)

	s.Equal(http.StatusOK, s.do(http.MethodGet, "/auth/roles/ADMIN", adminToken))
}

// burstLogin fires n token requests through the real app, applying mutate to
// each, and returns the status of the last one.
func (s *RoutesSuite) burstLogin(n int, mutate func(*http.Request)) int {
	body := fmt.Sprintf(`{"username":"admin","password":%q}`, adminPassword)
	last := 0
	for i := 0; i < n; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		if mutate != nil {
			mutate(req)
		}
		resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
		s.Require().NoError(err)
		last = resp.StatusCode
		_ = resp.Body.Close()
	}
	return last
}

// TestTokenEndpointIsRateLimited: with the default limit of 10 per window, the
// eleventh request in the window is rejected with 429 before the use case runs
// -- proved by the repository mocks expecting exactly ten calls.
func (s *RoutesSuite) TestTokenEndpointIsRateLimited() {
	s.userRepository.EXPECT().FindByUsername(gomock.Any(), "admin").
		Return(&entity.User{ID: s.userID, Username: "admin", Password: adminPasswordHash, Enabled: true}, nil).Times(10)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(gomock.Any(), s.userID).
		Return([]string{"USER"}, nil).Times(10)
	s.loginAttemptRepository.EXPECT().Get(gomock.Any(), s.userID).Return(nil, nil).Times(10)
	s.refreshTokenRepository.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(passthroughRefreshCreate).Times(10)

	s.Equal(http.StatusTooManyRequests, s.burstLogin(11, nil))
}

// TestForgedForwardedForDoesNotBypassRateLimit: TRUSTED_PROXIES is unset, so
// X-Forwarded-For is ignored and a new forged client address per request does
// not buy a fresh limiter bucket.
func (s *RoutesSuite) TestForgedForwardedForDoesNotBypassRateLimit() {
	s.userRepository.EXPECT().FindByUsername(gomock.Any(), "admin").
		Return(&entity.User{ID: s.userID, Username: "admin", Password: adminPasswordHash, Enabled: true}, nil).Times(10)
	s.userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(gomock.Any(), s.userID).
		Return([]string{"USER"}, nil).Times(10)
	s.loginAttemptRepository.EXPECT().Get(gomock.Any(), s.userID).Return(nil, nil).Times(10)
	s.refreshTokenRepository.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(passthroughRefreshCreate).Times(10)

	i := 0
	status := s.burstLogin(11, func(req *http.Request) {
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.0.0.%d", i))
		i++
	})
	s.Equal(http.StatusTooManyRequests, status)
}
