package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"golauth/model"
	"golauth/repository"
	"golauth/usecase"
	"golauth/util"
	"net/http"
)

var ErrContentTypeNotSuported = errors.New("content-type not supported")

type TokenController interface {
	Token(w http.ResponseWriter, r *http.Request)
}

type tokenController struct {
	userRepository          repository.UserRepository
	userAuthorityRepository repository.UserAuthorityRepository
	tokenService            usecase.TokenService
	userService             usecase.UserService
}

func NewTokenController(
	userRepository repository.UserRepository,
	userAuthorityRepository repository.UserAuthorityRepository,
	tokenService usecase.TokenService,
	userService usecase.UserService) TokenController {
	return tokenController{
		userRepository:          userRepository,
		userAuthorityRepository: userAuthorityRepository,
		tokenService:            tokenService,
		userService:             userService,
	}
}

func (s tokenController) Token(w http.ResponseWriter, r *http.Request) {
	var username string
	var password string
	var err error

	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		username, password, err = s.extractUserPasswordFromForm(r, username, password)
	} else if r.Header.Get("Content-Type") == "application/json" {
		username, password, err = s.extractUserPasswordFromJson(r, username, password)
	} else {
		util.SendBadRequest(w, ErrContentTypeNotSuported)
		return
	}

	if err != nil {
		util.SendServerError(w, err)
		return
	}
	tk, err := s.userService.GenerateToken(username, password)
	if err != nil {
		s.encapsulateModelError(w, err)
		return
	}

	util.SendSuccess(w, tk)
}

func (s tokenController) encapsulateModelError(w http.ResponseWriter, err error) {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusUnauthorized
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(e)
}

func (s tokenController) extractUserPasswordFromJson(r *http.Request, username string, password string) (string, string, error) {
	var userLogin model.UserLogin
	err := json.NewDecoder(r.Body).Decode(&userLogin)
	if err != nil {
		return "", "", fmt.Errorf("json decoder error: %s", err.Error())
	}
	username = userLogin.Username
	password = userLogin.Password
	return username, password, err
}

func (s tokenController) extractUserPasswordFromForm(r *http.Request, username string, password string) (string, string, error) {
	err := r.ParseForm()
	if err != nil {
		return "", "", fmt.Errorf("parse form error: %s", err.Error())
	}
	username = r.FormValue("username")
	password = r.FormValue("password")
	return username, password, nil
}
