package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"golauth/model"
	"golauth/repository"
	"golauth/util"
	"net/http"
)

type UserController struct {
	userRepository     repository.UserRepository
	userRoleRepository repository.UserRoleRepository
}

func NewUserController(uRepo repository.UserRepository, urRepo repository.UserRoleRepository) UserController {
	return UserController{
		userRepository:     uRepo,
		userRoleRepository: urRepo,
	}
}

func (u UserController) FindByUsername(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username, _ := params["username"]
	user, err := u.userRepository.FindByUsername(username)
	util.SendResult(w, user, err)
}

func (u UserController) AddRole(w http.ResponseWriter, r *http.Request) {
	var userRole model.UserRole
	_ = json.NewDecoder(r.Body).Decode(&userRole)
	savedUserRole, err := u.userRoleRepository.AddUserRole(userRole.UserID, userRole.RoleID)
	w.WriteHeader(http.StatusCreated)
	util.SendResult(w, savedUserRole, err)
}
