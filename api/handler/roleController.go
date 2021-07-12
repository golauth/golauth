package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	repository2 "golauth/infrastructure/repository"
	"golauth/model"
	"net/http"
	"strconv"
)

type RoleController struct {
	roleRepository repository2.RoleRepository
}

func NewRoleController(rRepo repository2.RoleRepository) RoleController {
	return RoleController{roleRepository: rRepo}
}

func (c RoleController) CreateRole(w http.ResponseWriter, r *http.Request) {
	var role model.Role
	_ = json.NewDecoder(r.Body).Decode(&role)
	data, err := c.roleRepository.Create(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(data)
}

func (c RoleController) EditRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	_, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "cannot cast [id] to [int]", http.StatusInternalServerError)
		return
	}
	var data model.Role
	_ = json.NewDecoder(r.Body).Decode(&data)
	err = c.roleRepository.Edit(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}
