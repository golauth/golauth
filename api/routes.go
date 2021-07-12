package api

import (
	"database/sql"
	"golauth/api/handler"
	"golauth/api/middleware"
	"golauth/infrastructure/repository"
	"golauth/usecase"
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
	tokenService         usecase.TokenService
}

func NewRouter(db *sql.DB) Router {
	uRepo := repository.NewUserRepository(db)
	rRepo := repository.NewRoleRepository(db)
	urRepo := repository.NewUserRoleRepository(db)
	uaRepo := repository.NewUserAuthorityRepository(db)
	tokenService := usecase.NewTokenService()
	userService := usecase.NewUserService(uRepo, rRepo, urRepo, uaRepo, tokenService)

	return &router{
		signupController:     handler.NewSignupController(userService),
		tokenController:      handler.NewTokenController(uRepo, uaRepo, tokenService, userService),
		checkTokenController: handler.NewCheckTokenController(tokenService),
		userController:       handler.NewUserController(uRepo, urRepo),
		roleController:       handler.NewRoleController(rRepo),
		tokenService:         tokenService,
	}
}

func (r *router) Config() *mux.Router {
	rt := mux.NewRouter().PathPrefix("/auth").Subrouter()
	rt.HandleFunc("/signup", r.signupController.CreateUser).Methods(http.MethodPost, http.MethodOptions).Name("signup")
	rt.HandleFunc("/token", r.tokenController.Token).Methods(http.MethodPost, http.MethodOptions).Name("token")
	rt.HandleFunc("/check_token", r.checkTokenController.CheckToken).Methods(http.MethodGet, http.MethodOptions).Name("checkToken")

	rt.HandleFunc("/users/{username}", r.userController.FindByUsername).Methods(http.MethodGet, http.MethodOptions).Name("getUser")
	rt.HandleFunc("/users/{username}/add-role", r.userController.AddRole).Methods(http.MethodPost, http.MethodOptions).Name("addRoleToUser")

	rt.HandleFunc("/roles", r.roleController.CreateRole).Methods(http.MethodPost, http.MethodOptions).Name("addRole")
	rt.HandleFunc("/roles/{id}", r.roleController.EditRole).Methods(http.MethodPut, http.MethodOptions).Name("editRole")

	rt.Use(middleware.NewCorsMiddleware("*").Apply)
	rt.Use(middleware.NewSecurityMiddleware(r.tokenService, pathPrefix).Apply)
	rt.Use(middleware.NewCommonMiddleware().Apply)

	return rt
}
