package controller

import (
	"database/sql"
	"encoding/json"
	"golauth/model"
	"golauth/repository"
	"golauth/util"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

const defaultRoleName = "USER"

type SignupController struct {
	userRepository     repository.UserRepository
	roleRepository     repository.RoleRepository
	userRoleRepository repository.UserRoleRepository
}

func NewSignupController(db *sql.DB) SignupController {
	return SignupController{
		userRepository:     repository.NewUserRepository(db),
		roleRepository:     repository.NewRoleRepository(db),
		userRoleRepository: repository.NewUserRoleRepository(db),
	}
}

func (s SignupController) CreateUser(w http.ResponseWriter, r *http.Request) {

	var decodedUser model.User
	_ = json.NewDecoder(r.Body).Decode(&decodedUser)

	hash, err := bcrypt.GenerateFromPassword([]byte(decodedUser.Password), bcrypt.DefaultCost)
	if err != nil {
		util.SendServerError(w, err)
		return
	}

	decodedUser.Password = string(hash)
	user, err := s.userRepository.Create(decodedUser)
	if err != nil {
		util.SendServerError(w, err)
		return
	}
	role, err := s.roleRepository.FindByName(defaultRoleName)
	if err != nil {
		util.SendServerError(w, err)
		return
	}
	_, err = s.userRoleRepository.AddUserRole(user.ID, role.ID)
	if err != nil {
		util.SendServerError(w, err)
		return
	}

	util.SendResult(w, user, err)
}
