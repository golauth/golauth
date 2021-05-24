package routes

import (
	"golauth/controller"
	"golauth/util"
	"net/http"

	"github.com/gorilla/mux"
)

type Routes struct {
	signUpController     controller.SignupController
	signInController     controller.SigninController
	checkTokenController controller.CheckTokenController
	userController       controller.UserController
	roleController       controller.RoleController
	publicURI            map[string]bool
}

//func init() {
//	signUpController = controller.SignupController{}
//	signInController = controller.SigninController{}
//	checkTokenController = controller.CheckTokenController{}
//	userController = controller.UserController{}
//	roleController = controller.RoleController{}
//	publicURI = map[string]bool{
//		"/golauth/token":       true,
//		"/golauth/check_token": true,
//		"/golauth/signup":      true,
//	}
//}

func NewRoutes(pathPrefix string) *Routes {
	return &Routes{
		signUpController:     controller.SignupController{},
		signInController:     controller.SigninController{},
		checkTokenController: controller.CheckTokenController{},
		userController:       controller.UserController{},
		roleController:       controller.RoleController{},
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
	router.HandleFunc("/users/{username}/add-role", r.userController.AddRole).Methods(http.MethodGet, http.MethodOptions)

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
			token, err := util.ExtractToken(request)
			if err != nil {
				util.SendError(responseWriter, err)
				return
			}
			err = util.ValidateToken(token)
			if err != nil {
				util.SendError(responseWriter, err)
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
