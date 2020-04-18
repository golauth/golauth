package controller

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"golauth/model"
	"golauth/repository"
	"golauth/util"
	"net/http"
	"strconv"
)

type RoleController struct{}

var roleRepository repository.RoleRepository

func init() {
	roleRepository = repository.RoleRepository{}
}

func (c RoleController) CreateRole(w http.ResponseWriter, r *http.Request) {
	var role model.Role
	_ = json.NewDecoder(r.Body).Decode(&role)
	savedRole, err := roleRepository.Create(role)
	util.SendResult(w, savedRole, err)
}

func (c RoleController) EditRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	_, err := strconv.Atoi(params["id"])
	if err != nil {
		util.SendServerError(w, errors.New("cannot cast [id] to [int]"))
		return
	}
	var role model.Role
	_ = json.NewDecoder(r.Body).Decode(&role)
	err = roleRepository.Edit(role)
	util.SendResult(w, role, err)
}