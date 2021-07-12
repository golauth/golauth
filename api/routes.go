package api

import (
	"database/sql"
	"golauth/api/handler"
	"golauth/api/middleware"
	"golauth/repository"
	"golauth/usecase"
	"net/http"

	"github.com/gorilla/mux"
)

type Router interface {
	RegisterRoutes(router *mux.Router)
}

type router struct {
	signupController     handler.SignupController
	tokenController      handler.TokenController
	checkTokenController handler.CheckTokenController
	userController       handler.UserController
	roleController       handler.RoleController
	tokenService         usecase.TokenService
	pathPrefix           string
}

func NewRouter(pathPrefix string, db *sql.DB) Router {
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
		pathPrefix:           pathPrefix,
	}
}

func (r *router) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/signup", r.signupController.CreateUser).Methods(http.MethodPost, http.MethodOptions).Name("signup")
	router.HandleFunc("/token", r.tokenController.Token).Methods(http.MethodPost, http.MethodOptions).Name("token")
	router.HandleFunc("/check_token", r.checkTokenController.CheckToken).Methods(http.MethodGet, http.MethodOptions).Name("checkToken")

	router.HandleFunc("/users/{username}", r.userController.FindByUsername).Methods(http.MethodGet, http.MethodOptions).Name("getUser")
	router.HandleFunc("/users/{username}/add-role", r.userController.AddRole).Methods(http.MethodPost, http.MethodOptions).Name("addRoleToUser")

	router.HandleFunc("/roles", r.roleController.CreateRole).Methods(http.MethodPost, http.MethodOptions).Name("addRole")
	router.HandleFunc("/roles/{id}", r.roleController.EditRole).Methods(http.MethodPut, http.MethodOptions).Name("editRole")

	router.Use(middleware.NewCorsMiddleware("*").Apply)
	router.Use(middleware.NewSecurityMiddleware(r.tokenService, r.pathPrefix).Apply)
	router.Use(middleware.NewCommonMiddleware().Apply)
}
