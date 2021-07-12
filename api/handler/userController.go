package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	repository2 "golauth/infrastructure/repository"
	"golauth/model"
	"net/http"
)

type UserController struct {
	userRepository     repository2.UserRepository
	userRoleRepository repository2.UserRoleRepository
}

func NewUserController(uRepo repository2.UserRepository, urRepo repository2.UserRoleRepository) UserController {
	return UserController{
		userRepository:     uRepo,
		userRoleRepository: urRepo,
	}
}

func (u UserController) FindByUsername(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username, _ := params["username"]
	data, err := u.userRepository.FindByUsername(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func (u UserController) AddRole(w http.ResponseWriter, r *http.Request) {
	var userRole model.UserRole
	_ = json.NewDecoder(r.Body).Decode(&userRole)
	data, err := u.userRoleRepository.AddUserRole(userRole.UserID, userRole.RoleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(data)
}
