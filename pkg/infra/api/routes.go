package api

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/application/user"
	"github.com/golauth/golauth/pkg/domain/factory"
	"github.com/golauth/golauth/pkg/infra/api/controller"
	"github.com/golauth/golauth/pkg/infra/api/middleware"
	"github.com/golauth/golauth/pkg/infra/keys"
	"github.com/sirupsen/logrus"
)

const pathPrefix = "/auth"

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
	validateToken        token.ValidateToken
}

func NewRouter(repoFactory factory.RepositoryFactory, keySet *keys.KeySet) Router {
	uRepo := repoFactory.NewUserRepository()
	urRepo := repoFactory.NewUserRoleRepository()
	uaRepo := repoFactory.NewUserAuthorityRepository()
	jwtToken := token.NewGenerateJwtToken(keySet.Current)

	createUser := user.NewCreateUser(repoFactory)
	findUserById := user.NewFindUserById(uRepo)
	addUserRole := user.NewAddUserRole(urRepo)
	generateToken := token.NewGenerateToken(repoFactory, jwtToken, lockoutPolicy())
	validateToken := token.NewValidateToken(keySet)

	return &router{
		signupController:     controller.NewSignupController(createUser),
		tokenController:      controller.NewTokenController(uRepo, uaRepo, generateToken),
		checkTokenController: controller.NewCheckTokenController(validateToken),
		userController:       controller.NewUserController(findUserById, addUserRole),
		roleController:       controller.NewRoleController(repoFactory),
		jwksController:       controller.NewJWKSController(keySet),
		validateToken:        validateToken,
	}
}

func (r *router) Config() *fiber.App {
	proxies := csvEnv("TRUSTED_PROXIES")
	app := fiber.New(fiber.Config{
		AppName: os.Getenv("APP_NAME"),
		// X-Forwarded-For is honoured only when the peer is one of these; without
		// this the header is attacker controlled and the login limiter, which
		// keys on client IP, is bypassed by simply sending a new value each time.
		TrustProxy:       len(proxies) > 0,
		TrustProxyConfig: fiber.TrustProxyConfig{Proxies: proxies},
	})

	// Middlewares are registered before any route on purpose. The fiber router
	// serves the first matching stack entry and stops, so a middleware added
	// after the routes never runs for them -- that is GHSA-p34g-m47x-q2m4.
	// Registering them first also makes a newly added route protected by
	// default: SecurityMiddleware opts paths out through its public allowlist,
	// never the other way around.
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins(),
		AllowMethods: []string{"POST", "GET", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"access-control-allow-headers", "access-control-allow-methods", "access-control-allow-origin", "authorization", "content-type"},
	}))
	app.Use(middleware.NewSecurityMiddleware(r.validateToken, pathPrefix).Apply())

	auth := app.Group(pathPrefix)

	// Public.
	auth.Post("/signup", r.signupController.CreateUser).Name("signup")
	// Deprecated: signup over GET carries the credentials in a request body
	// that proxies and access logs may retain. Removed in the next minor.
	auth.Get("/signup", r.signupController.CreateUser).Name("signupDeprecated")
	// The token route is the credential-stuffing surface, so it -- and only it
	// -- is rate limited per client IP. A busy authenticated API is untouched.
	auth.Post("/token", loginRateLimiter(), r.tokenController.Token).Name("token")
	auth.Get("/check_token", r.checkTokenController.CheckToken).Name("checkToken")
	// Public: the public signing keys, so any service can verify a token offline.
	auth.Get("/.well-known/jwks.json", r.jwksController.JWKS).Name("jwks")

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
	return limiter.New(limiter.Config{
		Max:               intEnv("LOGIN_RATE_LIMIT", defaultLoginRateLimit),
		Expiration:        durationEnv("LOGIN_RATE_WINDOW", defaultLoginRateWindow),
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator:      func(c fiber.Ctx) string { return c.IP() },
		LimitReached: func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many login attempts, retry later")
		},
	})
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
		logrus.Warnf("invalid %s=%q, using default %d", key, v, def)
	}
	return def
}

// durationEnv reads a positive Go duration, warning and using def on anything else.
func durationEnv(key string, def time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		logrus.Warnf("invalid %s=%q (want a Go duration like \"1m\"), using default %s", key, v, def)
	}
	return def
}
