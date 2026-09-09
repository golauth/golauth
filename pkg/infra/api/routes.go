package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/application/user"
	"github.com/golauth/golauth/pkg/domain/factory"
	"github.com/golauth/golauth/pkg/infra/api/controller"
	"github.com/golauth/golauth/pkg/infra/api/middleware"
	"os"
)

const pathPrefix = "/auth"

type Router interface {
	Config() *fiber.App
}

type router struct {
	signupController     controller.SignupController
	tokenController      controller.TokenController
	checkTokenController controller.CheckTokenController
	userController       controller.UserController
	roleController       controller.RoleController
	validateToken        token.ValidateToken
}

func NewRouter(repoFactory factory.RepositoryFactory) Router {
	uRepo := repoFactory.NewUserRepository()
	urRepo := repoFactory.NewUserRoleRepository()
	uaRepo := repoFactory.NewUserAuthorityRepository()
	key := token.GeneratePrivateKey()
	jwtToken := token.NewGenerateJwtToken(key)

	createUser := user.NewCreateUser(repoFactory)
	findUserById := user.NewFindUserById(uRepo)
	addUserRole := user.NewAddUserRole(urRepo)
	generateToken := token.NewGenerateToken(repoFactory, jwtToken)
	validateToken := token.NewValidateToken(key)

	return &router{
		signupController:     controller.NewSignupController(createUser),
		tokenController:      controller.NewTokenController(uRepo, uaRepo, generateToken),
		checkTokenController: controller.NewCheckTokenController(validateToken),
		userController:       controller.NewUserController(findUserById, addUserRole),
		roleController:       controller.NewRoleController(repoFactory),
		validateToken:        validateToken,
	}
}

func (r *router) Config() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               os.Getenv("APP_NAME"),
		DisableStartupMessage: true,
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
		AllowMethods: "POST, GET, OPTIONS, PUT, PATCH, DELETE",
		AllowHeaders: "access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,authorization,content-type",
	}))
	app.Use(middleware.NewSecurityMiddleware(r.validateToken, pathPrefix).Apply())

	auth := app.Group(pathPrefix)

	// Public.
	auth.Post("/signup", r.signupController.CreateUser).Name("signup")
	// Deprecated: signup over GET carries the credentials in a request body
	// that proxies and access logs may retain. Removed in the next minor.
	auth.Get("/signup", r.signupController.CreateUser).Name("signupDeprecated")
	auth.Post("/token", r.tokenController.Token).Name("token")
	auth.Get("/check_token", r.checkTokenController.CheckToken).Name("checkToken")

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

// allowedOrigins reads the browser origins allowed to call the API. It
// deliberately does not fall back to "*": combined with the authorization
// header being allowed, a wildcard lets any site drive the API with a token it
// coaxed out of a user.
func allowedOrigins() string {
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		return origins
	}
	return "http://localhost:3000"
}
