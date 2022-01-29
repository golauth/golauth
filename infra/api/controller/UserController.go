package controller

import (
	"encoding/json"
	"github.com/golauth/golauth/domain/usecase/user"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
)

type UserController struct {
	findById    user.FindUserById
	addUserRole user.AddUserRole
}

func NewUserController(findById user.FindUserById, addUserRole user.AddUserRole) UserController {
	return UserController{findById: findById, addUserRole: addUserRole}
}

func (u UserController) FindById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := u.findById.Execute(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func (u UserController) AddRole(w http.ResponseWriter, r *http.Request) {
	var userRole model.UserRoleRequest
	_ = json.NewDecoder(r.Body).Decode(&userRole)
	err := u.addUserRole.Execute(r.Context(), userRole.UserID, userRole.RoleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
