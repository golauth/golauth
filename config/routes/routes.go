package routes

import (
	"golauth/controller"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	signupController     controller.SignupController
	siginController      controller.SigninController
	checkTokenController controller.CheckTokenController
)

func init() {
	signupController = controller.SignupController{}
	siginController = controller.SigninController{}
	checkTokenController = controller.CheckTokenController{}
}

func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/signup", signupController.CreateUser).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/token", siginController.Token).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/check_token", checkTokenController.CheckToken).Methods(http.MethodGet)
	router.Use(configMiddleware)
}

func configMiddleware(next http.Handler) http.Handler {
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
