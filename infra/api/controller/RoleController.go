package controller

import (
	"encoding/json"
	"github.com/golauth/golauth/domain/factory"
	"github.com/golauth/golauth/domain/usecase/role"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
)

type RoleController struct {
	addRole          role.AddRole
	editRole         role.EditRole
	changeRoleStatus role.ChangeRoleStatus
	findByName       role.FindRoleByName
}

func NewRoleController(repoFactory factory.RepositoryFactory) RoleController {
	return RoleController{
		addRole:          role.NewAddRole(repoFactory),
		editRole:         role.NewEditRole(repoFactory.NewRoleRepository()),
		changeRoleStatus: role.NewChangeRoleStatus(repoFactory.NewRoleRepository()),
		findByName:       role.NewFindRoleByName(repoFactory.NewRoleRepository()),
	}
}

func (c RoleController) Create(w http.ResponseWriter, r *http.Request) {
	var roleRequest model.RoleRequest
	_ = json.NewDecoder(r.Body).Decode(&roleRequest)
	data, err := c.addRole.Execute(r.Context(), roleRequest)
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
	err = c.editRole.Execute(r.Context(), id, data)
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
	err = c.changeRoleStatus.Execute(r.Context(), id, data.Enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func (c RoleController) FindByName(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	data, err := c.findByName.Execute(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}
