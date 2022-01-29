package api

import (
	"github.com/golauth/golauth/domain/factory"
	"github.com/golauth/golauth/domain/usecase"
	"github.com/golauth/golauth/domain/usecase/token"
	"github.com/golauth/golauth/infra/api/controller"
	"github.com/golauth/golauth/infra/api/middleware"
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
	tokenService         token.UseCase
}

func NewRouter(repoFactory factory.RepositoryFactory) Router {
	uRepo := repoFactory.NewUserRepository()
	rRepo := repoFactory.NewRoleRepository()
	urRepo := repoFactory.NewUserRoleRepository()
	uaRepo := repoFactory.NewUserAuthorityRepository()
	tokenService := token.NewService()
	userService := usecase.NewUserService(uRepo, rRepo, urRepo, uaRepo, tokenService)

	roleSvc := usecase.NewRoleService(rRepo)

	return &router{
		signupController:     controller.NewSignupController(userService),
		tokenController:      controller.NewTokenController(uRepo, uaRepo, tokenService, userService),
		checkTokenController: controller.NewCheckTokenController(tokenService),
		userController:       controller.NewUserController(userService),
		roleController:       controller.NewRoleController(roleSvc, repoFactory),
		tokenService:         tokenService,
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
	rt.Use(middleware.NewSecurityMiddleware(r.tokenService, pathPrefix).Apply)
	rt.Use(middleware.NewCommonMiddleware().Apply)

	return rt
}
