package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"golauth/model"
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

func (u UserController) AddRole(w http.ResponseWriter, r *http.Request) {
	var userRole model.UserRole
	_ = json.NewDecoder(r.Body).Decode(&userRole)
	userRoleRepository := repository.UserRoleRepository{}
	savedUserRole, err := userRoleRepository.AddUserRole(userRole.UserID, userRole.RoleID)
	util.SendResult(w, savedUserRole, err)
}
