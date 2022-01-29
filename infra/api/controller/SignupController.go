package controller

import (
	"encoding/json"
	"github.com/golauth/golauth/domain/usecase/user"
	"github.com/golauth/golauth/infra/api/controller/model"
	"net/http"
)

type SignupController interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
}

type signupController struct {
	createUser user.CreateUser
}

func NewSignupController(createUser user.CreateUser) SignupController {
	return signupController{createUser: createUser}
}

func (s signupController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var decodedUser model.UserRequest
	_ = json.NewDecoder(r.Body).Decode(&decodedUser)
	output, err := s.createUser.Execute(r.Context(), decodedUser.ToEntity())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(model.NewUserResponseFromEntity(output))
}
