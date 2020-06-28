package routes

import (
	"golauth/controller"
	"golauth/util"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	signupController     controller.SignupController
	siginController      controller.SigninController
	checkTokenController controller.CheckTokenController
	userController       controller.UserController
	roleController       controller.RoleController
	publicURI            map[string]bool
)

func init() {
	signupController = controller.SignupController{}
	siginController = controller.SigninController{}
	checkTokenController = controller.CheckTokenController{}
	userController = controller.UserController{}
	roleController = controller.RoleController{}
	publicURI = map[string]bool{
		"/golauth/token":       true,
		"/golauth/check_token": true,
		"/golauth/signup":      true,
	}
}

func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/signup", signupController.CreateUser).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/token", siginController.Token).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/check_token", checkTokenController.CheckToken).Methods(http.MethodGet, http.MethodOptions)

	router.HandleFunc("/users/{username}", userController.FindByUsername).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/users/{username}/add-role", userController.AddRole).Methods(http.MethodGet, http.MethodOptions)

	router.HandleFunc("/roles", roleController.CreateRole).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/roles/{id}", roleController.EditRole).Methods(http.MethodPut, http.MethodOptions)
	router.Use(applyCors)
	router.Use(applySecurity)
}

func applyCors(next http.Handler) http.Handler {
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

func applySecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURI := r.RequestURI
		if isPrivateURI(requestURI) {
			token, err := util.ExtractToken(r)
			if err != nil {
				util.SendError(w, err)
				return
			}
			err = util.ValidateToken(token)
			if err != nil {
				util.SendError(w, err)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func isPrivateURI(requestURI string) bool {
	_, contains := publicURI[requestURI]
	return !contains
}
