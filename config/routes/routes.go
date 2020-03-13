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
	router.HandleFunc("/signup", signupController.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/token", siginController.Token).Methods(http.MethodPost)
	router.HandleFunc("/check_token", checkTokenController.CheckToken).Methods(http.MethodGet)
}
