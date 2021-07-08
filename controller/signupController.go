package controller

import (
	"encoding/json"
	"golauth/model"
	"golauth/usecase"
	"golauth/util"
	"net/http"
)

type SignupController interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
}

type signupController struct {
	service usecase.SignupService
}

func NewSignupController(service usecase.SignupService) SignupController {
	return signupController{service: service}
}

func (s signupController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var decodedUser model.User
	_ = json.NewDecoder(r.Body).Decode(&decodedUser)
	user, err := s.service.CreateUser(decodedUser)
	w.WriteHeader(http.StatusCreated)
	util.SendResult(w, user, err)
}
