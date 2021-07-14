package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golauth/model"
	"golauth/usecase"
	"net/http"
)

type UserController struct {
	userSvc usecase.UserService
}

func NewUserController(uSvc usecase.UserService) UserController {
	return UserController{userSvc: uSvc}
}

func (u UserController) FindById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := u.userSvc.FindByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func (u UserController) AddRole(w http.ResponseWriter, r *http.Request) {
	var userRole model.UserRoleRequest
	_ = json.NewDecoder(r.Body).Decode(&userRole)
	err := u.userSvc.AddUserRole(userRole.UserID, userRole.RoleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
