package routes

import (
	"database/sql"
	"golauth/controller"
	"golauth/model"
	"golauth/repository"
	"golauth/usecase"
	"golauth/util"
	"net/http"

	"github.com/gorilla/mux"
)

type Routes struct {
	signUpController     controller.SignupController
	signInController     controller.SignInController
	checkTokenController controller.CheckTokenController
	userController       controller.UserController
	roleController       controller.RoleController
	tokenService         usecase.TokenService
	publicURI            map[string]bool
}

func NewRoutes(pathPrefix string, db *sql.DB, privBytes []byte, pubBytes []byte) *Routes {
	tokenService := usecase.NewTokenService(privBytes, pubBytes)
	uRepo := repository.NewUserRepository(db)
	rRepo := repository.NewRoleRepository(db)
	urRepo := repository.NewUserRoleRepository(db)
	signupService := usecase.NewSignupService(uRepo, rRepo, urRepo)

	return &Routes{
		signUpController:     controller.NewSignupController(signupService),
		signInController:     controller.NewSignInController(db, privBytes, pubBytes),
		checkTokenController: controller.NewCheckTokenController(tokenService),
		userController:       controller.NewUserController(uRepo, urRepo),
		roleController:       controller.NewRoleController(rRepo),
		tokenService:         tokenService,
		publicURI: map[string]bool{
			pathPrefix + "/token":       true,
			pathPrefix + "/check_token": true,
			pathPrefix + "/signup":      true,
		},
	}
}

func (r *Routes) RegisterRouter(router *mux.Router) {
	router.HandleFunc("/signup", r.signUpController.CreateUser).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/token", r.signInController.Token).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/check_token", r.checkTokenController.CheckToken).Methods(http.MethodGet, http.MethodOptions)

	router.HandleFunc("/users/{username}", r.userController.FindByUsername).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/users/{username}/add-role", r.userController.AddRole).Methods(http.MethodPost, http.MethodOptions)

	router.HandleFunc("/roles", r.roleController.CreateRole).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/roles/{id}", r.roleController.EditRole).Methods(http.MethodPut, http.MethodOptions)
	router.Use(r.applyCors)
	router.Use(r.applySecurity)
}

func (r *Routes) applyCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,authorization")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (r *Routes) applySecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		requestURI := request.RequestURI
		if r.isPrivateURI(requestURI) {
			token, err := r.tokenService.ExtractToken(request)
			if err != nil {
				util.SendError(responseWriter, &model.Error{StatusCode: http.StatusBadGateway, Message: err.Error()})
				return
			}
			err = r.tokenService.ValidateToken(token)
			if err != nil {
				util.SendError(responseWriter, &model.Error{StatusCode: http.StatusUnauthorized, Message: err.Error()})
				return
			}
		}
		next.ServeHTTP(responseWriter, request)
	})
}

func (r *Routes) isPrivateURI(requestURI string) bool {
	_, contains := r.publicURI[requestURI]
	return !contains
}
