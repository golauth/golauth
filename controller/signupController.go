package controller

import (
	"encoding/json"
	"golauth/model"
	"golauth/repository"
	"golauth/util"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type SignupController struct{}

var defaultRoleName string

func init() {
	defaultRoleName = "USER"
}

func (s SignupController) CreateUser(w http.ResponseWriter, r *http.Request) {
	userRespository := repository.UserRepository{}
	roleRepository := repository.RoleRepository{}
	userRoleRepository := repository.UserRoleRepository{}
	var decodedUser model.User
	_ = json.NewDecoder(r.Body).Decode(&decodedUser)

	hash, err := bcrypt.GenerateFromPassword([]byte(decodedUser.Password), bcrypt.DefaultCost)
	if err != nil {
		util.SendServerError(w, err)
		return
	}

	decodedUser.Password = string(hash)
	user, err := userRespository.Create(decodedUser)
	if err != nil {
		util.SendServerError(w, err)
		return
	}
	role, err := roleRepository.FindByName(defaultRoleName)
	if err != nil {
		util.SendServerError(w, err)
		return
	}
	_, err = userRoleRepository.AddUserRole(user.ID, role.ID)
	if err != nil {
		util.SendServerError(w, err)
		return
	}

	util.SendResult(w, user, err)
}
