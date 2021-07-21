package api

import (
	"database/sql"
	"golauth/api/handler"
	"golauth/api/middleware"
	"golauth/infrastructure/repository"
	"golauth/usecase"
	"golauth/usecase/token"
	"net/http"

	"github.com/gorilla/mux"
)

const pathPrefix = "/auth"

type Router interface {
	Config() *mux.Router
}

type router struct {
	signupController     handler.SignupController
	tokenController      handler.TokenController
	checkTokenController handler.CheckTokenController
	userController       handler.UserController
	roleController       handler.RoleController
	tokenService         token.UseCase
}

func NewRouter(db *sql.DB) Router {
	uRepo := repository.NewUserRepository(db)
	rRepo := repository.NewRoleRepository(db)
	urRepo := repository.NewUserRoleRepository(db)
	uaRepo := repository.NewUserAuthorityRepository(db)
	tokenService := token.NewService()
	userService := usecase.NewUserService(uRepo, rRepo, urRepo, uaRepo, tokenService)

	roleSvc := usecase.NewRoleService(rRepo)

	return &router{
		signupController:     handler.NewSignupController(userService),
		tokenController:      handler.NewTokenController(uRepo, uaRepo, tokenService, userService),
		checkTokenController: handler.NewCheckTokenController(tokenService),
		userController:       handler.NewUserController(userService),
		roleController:       handler.NewRoleController(roleSvc),
		tokenService:         tokenService,
	}
}

func (r *router) Config() *mux.Router {
	rt := mux.NewRouter().PathPrefix("/auth").Subrouter()

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
