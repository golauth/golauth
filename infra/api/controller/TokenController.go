package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/domain/usecase/token"
	"github.com/golauth/golauth/infra/api/controller/model"
	"net/http"
)

var ErrContentTypeNotSupported = errors.New("content-type not supported")

type TokenController interface {
	Token(w http.ResponseWriter, r *http.Request)
}

type tokenController struct {
	userRepository          repository.UserRepository
	userAuthorityRepository repository.UserAuthorityRepository
	generateToken           token.GenerateToken
	tokenService            token.UseCase
}

func NewTokenController(
	userRepository repository.UserRepository,
	userAuthorityRepository repository.UserAuthorityRepository,
	tokenService token.UseCase,
	generateToken token.GenerateToken) TokenController {
	return tokenController{
		userRepository:          userRepository,
		userAuthorityRepository: userAuthorityRepository,
		generateToken:           generateToken,
		tokenService:            tokenService,
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
		http.Error(w, ErrContentTypeNotSupported.Error(), http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := s.generateToken.Execute(r.Context(), username, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(w).Encode(data)
}

func (s tokenController) extractUserPasswordFromJson(r *http.Request, username string, password string) (string, string, error) {
	var userLogin model.UserLoginRequest
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
