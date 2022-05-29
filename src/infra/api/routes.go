package api

import (
	"github.com/golauth/golauth/src/application/token"
	"github.com/golauth/golauth/src/application/user"
	"github.com/golauth/golauth/src/domain/factory"
	"github.com/golauth/golauth/src/infra/api/controller"
	"github.com/golauth/golauth/src/infra/api/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

const pathPrefix = "/auth"

type Router interface {
	Config() *mux.Router
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

func (r *router) Config() *mux.Router {
	rt := mux.NewRouter().PathPrefix(pathPrefix).Subrouter()

	rt.HandleFunc("/signup", r.signupController.CreateUser).Methods(http.MethodPost, http.MethodOptions).Name("signup")
	rt.HandleFunc("/token", r.tokenController.Token).Methods(http.MethodPost, http.MethodOptions).Name("token")
	rt.HandleFunc("/check_token", r.checkTokenController.CheckToken).Methods(http.MethodGet, http.MethodOptions).Name("checkToken")

	rt.HandleFunc("/users/{id}", r.userController.FindById).Methods(http.MethodGet, http.MethodOptions).Name("getUser")
	rt.HandleFunc("/users/{id}/add-role", r.userController.AddRole).Methods(http.MethodPost, http.MethodOptions).Name("addRoleToUser")

	rt.HandleFunc("/roles", r.roleController.Create).Methods(http.MethodPost, http.MethodOptions).Name("addRole")
	rt.HandleFunc("/roles/{name}", r.roleController.FindByName).Methods(http.MethodGet, http.MethodOptions).Name("findRoleByName")
	rt.HandleFunc("/roles/{id}", r.roleController.Edit).Methods(http.MethodPut, http.MethodOptions).Name("editRole")
	rt.HandleFunc("/roles/{id}/change-status", r.roleController.ChangeStatus).Methods(http.MethodPatch, http.MethodOptions).Name("changeStatus")

	rt.Use(middleware.NewCorsMiddleware("*").Apply)
	rt.Use(middleware.NewSecurityMiddleware(r.validateToken, pathPrefix).Apply)
	rt.Use(middleware.NewCommonMiddleware().Apply)

	return rt
}