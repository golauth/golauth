package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golauth/golauth/src/application/token"
	"github.com/golauth/golauth/src/application/user"
	"github.com/golauth/golauth/src/domain/factory"
	"github.com/golauth/golauth/src/infra/api/controller"
	"github.com/golauth/golauth/src/infra/api/middleware"
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
	app := fiber.New()

	auth := app.Group(pathPrefix)

	auth.Get("/signup", r.signupController.CreateUser).Name("signup")
	auth.Post("/token", r.tokenController.Token).Name("token")
	auth.Get("/check_token", r.checkTokenController.CheckToken).Name("checkToken")

	auth.Get("/users/:id", r.userController.FindById).Name("getUser")
	auth.Post("/users/:id/add-role", r.userController.AddRole).Name("addRoleToUser")

	auth.Post("/roles", r.roleController.Create).Name("addRole")
	auth.Get("/roles/:name", r.roleController.FindByName).Name("findRoleByName")
	auth.Put("/roles/:id", r.roleController.Edit).Name("editRole")
	auth.Patch("/roles/:id/change-status", r.roleController.ChangeStatus).Name("changeStatus")

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "POST, GET, OPTIONS, PUT, PATCH, DELETE",
		AllowHeaders: "access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,authorization",
	}))
	app.Use(middleware.NewSecurityMiddleware(r.validateToken, pathPrefix).Apply())
	app.Use(recover.New())

	return app
}
