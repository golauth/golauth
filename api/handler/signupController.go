package handler

import (
	"encoding/json"
	"golauth/model"
	"golauth/usecase"
	"net/http"
)

type SignupController interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
}

type signupController struct {
	service usecase.UserService
}

func NewSignupController(service usecase.UserService) SignupController {
	return signupController{service: service}
}

func (s signupController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var decodedUser model.User
	_ = json.NewDecoder(r.Body).Decode(&decodedUser)
	data, err := s.service.CreateUser(decodedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(data)
}
