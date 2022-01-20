package controller

import (
	"encoding/json"
	"github.com/golauth/golauth/domain/usecase"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
)

type RoleController struct {
	svc usecase.RoleService
}

func NewRoleController(s usecase.RoleService) RoleController {
	return RoleController{svc: s}
}

func (c RoleController) Create(w http.ResponseWriter, r *http.Request) {
	var role model.RoleRequest
	_ = json.NewDecoder(r.Body).Decode(&role)
	data, err := c.svc.Create(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(data)
}

func (c RoleController) Edit(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "cannot cast [id] to [uuid]", http.StatusInternalServerError)
		return
	}
	var data model.RoleRequest
	_ = json.NewDecoder(r.Body).Decode(&data)
	err = c.svc.Edit(id, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func (c RoleController) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "cannot cast [id] to [uuid]", http.StatusBadRequest)
		return
	}
	var data model.RoleChangeStatus
	_ = json.NewDecoder(r.Body).Decode(&data)
	err = c.svc.ChangeStatus(id, data.Enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func (c RoleController) FindByName(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	data, err := c.svc.FindByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}
