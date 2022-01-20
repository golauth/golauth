package handler

import (
	"encoding/json"
	"golauth/domain/usecase"
	"golauth/model"
	"net/http"
)

type SignupController interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
}

type signupController struct {
	svc usecase.UserService
}

func NewSignupController(s usecase.UserService) SignupController {
	return signupController{svc: s}
}

func (s signupController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var decodedUser model.UserRequest
	_ = json.NewDecoder(r.Body).Decode(&decodedUser)
	data, err := s.svc.CreateUser(decodedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(data)
}
