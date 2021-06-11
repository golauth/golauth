package controller

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"golauth/model"
	"golauth/repository"
	"golauth/util"
	"net/http"
)

type UserController struct {
	userRepository     repository.UserRepository
	userRoleRepository repository.UserRoleRepository
}

func NewUserController(db *sql.DB) UserController {
	return UserController{
		userRepository:     repository.NewUserRepository(db),
		userRoleRepository: repository.NewUserRoleRepository(db),
	}
}

func (u UserController) FindByUsername(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username, _ := params["username"]
	user, err := u.userRepository.FindByUsername(username)
	util.SendResult(w, user, err)
}

func (u UserController) AddRole(w http.ResponseWriter, r *http.Request) {
	var userRole model.UserRole
	_ = json.NewDecoder(r.Body).Decode(&userRole)
	savedUserRole, err := u.userRoleRepository.AddUserRole(userRole.UserID, userRole.RoleID)
	util.SendResult(w, savedUserRole, err)
}
