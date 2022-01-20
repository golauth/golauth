package handler

import (
	"encoding/json"
	"github.com/golauth/golauth/api/handler/model"
	"github.com/golauth/golauth/domain/usecase"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
)

type UserController struct {
	svc usecase.UserService
}

func NewUserController(s usecase.UserService) UserController {
	return UserController{svc: s}
}

func (u UserController) FindById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := u.svc.FindByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func (u UserController) AddRole(w http.ResponseWriter, r *http.Request) {
	var userRole model.UserRoleRequest
	_ = json.NewDecoder(r.Body).Decode(&userRole)
	err := u.svc.AddUserRole(userRole.UserID, userRole.RoleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
