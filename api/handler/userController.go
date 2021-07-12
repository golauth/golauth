package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"golauth/model"
	"golauth/repository"
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
