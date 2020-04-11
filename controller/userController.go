package controller

import (
	"github.com/gorilla/mux"
	"golauth/repository"
	"golauth/util"
	"net/http"
)

type UserController struct{}

func (u UserController) FindByUsername(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username, _ := params["username"]
	userRespository := repository.UserRepository{}
	user, err := userRespository.FindByUsername(username)
	util.SendResult(w, user, err)
}
