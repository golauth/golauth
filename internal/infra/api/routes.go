package api

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/golauth/golauth/internal/application/keys"
	"github.com/golauth/golauth/internal/application/token"
	"github.com/golauth/golauth/internal/application/user"
	"github.com/golauth/golauth/internal/domain/factory"
	"github.com/golauth/golauth/internal/infra/api/controller"
	"github.com/golauth/golauth/internal/infra/api/httperr"
	"github.com/golauth/golauth/internal/infra/api/middleware"
	"github.com/golauth/golauth/internal/infra/database"
)

const pathPrefix = "/auth"

// Health probe paths. They sit outside pathPrefix and are opened explicitly in
// the SecurityMiddleware allowlist.
const (
	livePath  = "/health/live"
	readyPath = "/health/ready"
)

// Server-level limits. A slow or oversized client must not be able to hold a
// connection open indefinitely, and no endpoint accepts more than a small JSON
// document. Each is overridable through the environment.
const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second
	defaultBodyLimit    = 64 * 1024
)

// defaultAllowedOrigin is used when CORS_ALLOWED_ORIGINS names no usable origin.
const defaultAllowedOrigin = "http://localhost:3000"

const (
	defaultLoginRateLimit  = 10
	defaultLoginRateWindow = time.Minute
)

type Router interface {
	Config() *fiber.App
}

type router struct {
	signupController     controller.SignupController
	tokenController      controller.TokenController
	checkTokenController controller.CheckTokenController
	userController       controller.UserController
	roleController       controller.RoleController
	jwksController       controller.JWKSController
	healthController     controller.HealthController
	validateToken        token.ValidateToken
}

func NewRouter(repoFactory factory.RepositoryFactory, keySet *keys.KeySet, db database.Database) Router {
	uRepo := repoFactory.NewUserRepository()
	urRepo := repoFactory.NewUserRoleRepository()
	uaRepo := repoFactory.NewUserAuthorityRepository()
	tokenCfg := tokenConfig()
	jwtToken := token.NewGenerateJwtToken(keySet.Current, tokenCfg.AccessTokenTTL)

	createUser := user.NewCreateUser(repoFactory, user.LoadPasswordDenylist())
	findUserById := user.NewFindUserById(uRepo)
	addUserRole := user.NewAddUserRole(urRepo)
	generateToken := token.NewGenerateToken(repoFactory, jwtToken, lockoutPolicy(), tokenCfg)
	refreshAccessToken := token.NewRefreshAccessToken(repoFactory, jwtToken, tokenCfg)
	logout := token.NewLogout(repoFactory)
	validateToken := token.NewValidateToken(keySet)

	keyLoaded := func() bool {
		return keySet != nil && keySet.Current != nil && keySet.Current.Private != nil
	}

	return &router{
		signupController:     controller.NewSignupController(createUser),
		tokenController:      controller.NewTokenController(uRepo, uaRepo, generateToken, refreshAccessToken, logout),
		checkTokenController: controller.NewCheckTokenController(validateToken),
		userController:       controller.NewUserController(findUserById, addUserRole),
		roleController:       controller.NewRoleController(repoFactory),
		jwksController:       controller.NewJWKSController(keySet),
		healthController:     controller.NewHealthController(db, keyLoaded),
		validateToken:        validateToken,
	}
}

func (r *router) Config() *fiber.App {
	proxies := csvEnv("TRUSTED_PROXIES")
	app := fiber.New(fiber.Config{
		AppName: os.Getenv("APP_NAME"),
		// One error contract for the whole API: every handler and middleware
		// returns a plain error and this decides the status and the body.
		ErrorHandler: httperr.Handler,
		// X-Forwarded-For is honoured only when the peer is one of these; without
		// this the header is attacker controlled and the login limiter, which
		// keys on client IP, is bypassed by simply sending a new value each time.
		TrustProxy:       len(proxies) > 0,
		TrustProxyConfig: fiber.TrustProxyConfig{Proxies: proxies},
		// Explicit server limits: a slow client cannot hold a connection open
		// past ReadTimeout, and a request body is capped well under the 4 MB
		// default since every endpoint takes only a small JSON document.
		ReadTimeout:  durationEnv("SERVER_READ_TIMEOUT", defaultReadTimeout),
		WriteTimeout: durationEnv("SERVER_WRITE_TIMEOUT", defaultWriteTimeout),
		IdleTimeout:  durationEnv("SERVER_IDLE_TIMEOUT", defaultIdleTimeout),
		BodyLimit:    intEnv("SERVER_BODY_LIMIT", defaultBodyLimit),
	})

	// Middlewares are registered before any route on purpose. The fiber router
	// serves the first matching stack entry and stops, so a middleware added
	// after the routes never runs for them -- that is GHSA-p34g-m47x-q2m4.
	// Registering them first also makes a newly added route protected by
	// default: SecurityMiddleware opts paths out through its public allowlist,
	// never the other way around.
	//
	// RequestID is first so even a panic or an auth rejection carries a
	// correlation id into the error body and the log. AccessLog is next, ahead
	// of recover and the security middleware, so every request -- including a
	// 401 from SecurityMiddleware and a 500 from a recovered panic -- produces
	// exactly one structured line with that id.
	app.Use(middleware.RequestID())
	app.Use(middleware.AccessLog())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins(),
		AllowMethods: []string{"POST", "GET", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"access-control-allow-headers", "access-control-allow-methods", "access-control-allow-origin", "authorization", "content-type"},
	}))
	app.Use(middleware.NewSecurityMiddleware(r.validateToken, pathPrefix).Apply())

	// Health probes: public (opened in the SecurityMiddleware allowlist) and
	// kept out of the access log so a probe every few seconds is not noise.
	// Liveness never touches a dependency; readiness pings the database.
	app.Get(livePath, r.healthController.Live).Name("healthLive")
	app.Get(readyPath, r.healthController.Ready).Name("healthReady")

	auth := app.Group(pathPrefix)

	// Public. Signup is POST only: it creates a persistent account, so a safe,
	// retriable, prefetchable verb is the wrong shape, and credentials in a GET
	// (body or query string) leak into logs, history and proxies.
	auth.Post("/signup", r.signupController.CreateUser).Name("signup")
	// The token route is the credential-stuffing surface, so it -- and only it
	// -- is rate limited per client IP. A busy authenticated API is untouched.
	auth.Post("/token", loginRateLimiter(), r.tokenController.Token).Name("token")
	// Public: exchange an opaque refresh token for a fresh access/refresh pair.
	auth.Post("/token/refresh", r.tokenController.Refresh).Name("refreshToken")
	// Public and does public-key crypto per call, so it is rate limited per
	// client IP like the token route.
	auth.Get("/check_token", publicCryptoRateLimiter(), r.checkTokenController.CheckToken).Name("checkToken")
	// Public: the public signing keys, so any service can verify a token offline.
	auth.Get("/.well-known/jwks.json", r.jwksController.JWKS).Name("jwks")

	// Authenticated: a caller revokes its own sessions. Any valid access token
	// is enough -- no special authority.
	auth.Post("/logout", r.tokenController.Logout).Name("logout")
	auth.Post("/logout/all", r.tokenController.LogoutAll).Name("logoutAll")
	// Authenticated: return the caller's own verified claims, no database.
	auth.Get("/me", r.checkTokenController.Me).Name("me")

	// Authenticated, and authorized per route.
	auth.Get("/users/:id",
		middleware.RequireSelfOrAuthority("id", middleware.AdminAuthority),
		r.userController.FindById).Name("getUser")
	auth.Post("/users/:id/add-role",
		middleware.RequireAuthority(middleware.AdminAuthority),
		r.userController.AddRole).Name("addRoleToUser")

	admin := middleware.RequireAuthority(middleware.AdminAuthority)
	auth.Post("/roles", admin, r.roleController.Create).Name("addRole")
	auth.Get("/roles/:name", admin, r.roleController.FindByName).Name("findRoleByName")
	auth.Put("/roles/:id", admin, r.roleController.Edit).Name("editRole")
	auth.Patch("/roles/:id/change-status", admin, r.roleController.ChangeStatus).Name("changeStatus")

	return app
}

// loginRateLimiter throttles POST /auth/token per client IP. The numbers come
// from LOGIN_RATE_LIMIT (requests) and LOGIN_RATE_WINDOW (a Go duration such as
// "1m"); a sliding window avoids the burst-at-the-boundary weakness of a fixed
// one.
func loginRateLimiter() fiber.Handler {
	return ipRateLimiter("too many login attempts, retry later")
}

// publicCryptoRateLimiter throttles GET /auth/check_token per client IP. It is
// public and runs public-key verification on every call, so it must not be a
// free signature oracle. It shares the LOGIN_RATE_* knobs and keeps its own
// per-IP budget.
func publicCryptoRateLimiter() fiber.Handler {
	return ipRateLimiter("too many requests, retry later")
}

func ipRateLimiter(limitReachedMsg string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:               intEnv("LOGIN_RATE_LIMIT", defaultLoginRateLimit),
		Expiration:        durationEnv("LOGIN_RATE_WINDOW", defaultLoginRateWindow),
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator:      func(c fiber.Ctx) string { return c.IP() },
		LimitReached: func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, limitReachedMsg)
		},
	})
}

// tokenConfig reads the token lifetimes. ACCESS_TOKEN_TTL and REFRESH_TOKEN_TTL
// are Go durations; unset or invalid values fall back to the token package
// defaults (15m and 7 days).
func tokenConfig() token.Config {
	return token.Config{
		AccessTokenTTL:  durationEnv("ACCESS_TOKEN_TTL", token.DefaultAccessTokenTTL),
		RefreshTokenTTL: durationEnv("REFRESH_TOKEN_TTL", token.DefaultRefreshTokenTTL),
	}
}

// lockoutPolicy reads the per-account lockout knobs. Unset or invalid values
// fall back to token.DefaultLockoutPolicy.
func lockoutPolicy() token.LockoutPolicy {
	return token.LockoutPolicy{
		Threshold: intEnv("LOGIN_LOCKOUT_THRESHOLD", token.DefaultLockoutPolicy.Threshold),
		BaseDelay: durationEnv("LOGIN_LOCKOUT_BASE_DELAY", token.DefaultLockoutPolicy.BaseDelay),
		MaxDelay:  durationEnv("LOGIN_LOCKOUT_MAX_DELAY", token.DefaultLockoutPolicy.MaxDelay),
	}
}

// allowedOrigins reads the browser origins allowed to call the API. It
// deliberately does not fall back to "*": combined with the authorization
// header being allowed, a wildcard lets any site drive the API with a token it
// coaxed out of a user.
// Returning an empty slice would be that wildcard: fiber v3 reads no configured
// origin as permission to allow every one of them.
func allowedOrigins() []string {
	// The variable stays comma separated, as it was under fiber v2, while the
	// middleware now takes a slice.
	allowed := csvEnv("CORS_ALLOWED_ORIGINS")
	if len(allowed) == 0 {
		return []string{defaultAllowedOrigin}
	}
	return allowed
}

// csvEnv splits a comma-separated variable, trimming blanks.
func csvEnv(key string) []string {
	var out []string
	for _, part := range strings.Split(os.Getenv(key), ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

// intEnv reads a positive integer, warning and using def on anything else.
func intEnv(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
		slog.Warn("invalid env var, using default", "key", key, "value", v, "default", def)
	}
	return def
}

// durationEnv reads a positive Go duration, warning and using def on anything else.
func durationEnv(key string, def time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		slog.Warn("invalid duration env var, using default",
			"key", key, "value", v, "want", "a Go duration like \"1m\"", "default", def.String())
	}
	return def
}
