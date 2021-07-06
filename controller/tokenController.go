package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golauth/model"
	"golauth/repository"
	"golauth/usercase"
	"golauth/util"
	"net/http"
)

type SignInController interface {
	Token(w http.ResponseWriter, r *http.Request)
}

type signInController struct {
	userRepository          repository.UserRepository
	userAuthorityRepository repository.UserAuthorityRepository
	tokenService            usercase.TokenService
}

func NewSignInController(db *sql.DB, privBytes []byte, pubBytes []byte) SignInController {
	return signInController{
		userRepository:          repository.NewUserRepository(db),
		userAuthorityRepository: repository.NewUserAuthorityRepository(db),
		tokenService:            usercase.NewTokenService(privBytes, pubBytes),
	}
}

func (s signInController) Token(w http.ResponseWriter, r *http.Request) {
	var username string
	var password string
	var err error

	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		username, password, err = s.extractUserPassForm(r, username, password)
	} else if r.Header.Get("Content-Type") == "application/json" {
		username, password, err = s.extractUserPassJson(r, username, password)
	} else {
		util.SendBadRequest(w, errors.New("Content-Type not supported"))
		return
	}

	if err != nil {
		util.SendServerError(w, err)
		return
	}

	user, err := s.userRepository.FindByUsernameWithPassword(username)
	if (model.User{}) == user || &user == nil {
		e := s.usernameNotFoundError(w)
		_ = json.NewEncoder(w).Encode(e)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		e := s.invalidPasswordError(w, err)
		_ = json.NewEncoder(w).Encode(e)
		return
	}

	authorities, _ := s.loadAuthorities(user.ID)

	jwtToken, err := s.tokenService.GenerateJwtToken(user, authorities)
	if err != nil {
		e := s.tokenError(w, err)
		_ = json.NewEncoder(w).Encode(e)
		return
	}
	tokenResponse := model.TokenResponse{AccessToken: jwtToken}
	util.SendSuccess(w, tokenResponse)
}

func (s signInController) extractUserPassJson(r *http.Request, username string, password string) (string, string, error) {
	var userLogin model.UserLogin
	err := json.NewDecoder(r.Body).Decode(&userLogin)
	if err != nil {
		return "", "", fmt.Errorf("json decoder error: %s", err.Error())
	}
	username = userLogin.Username
	password = userLogin.Password
	return username, password, err
}

func (s signInController) extractUserPassForm(r *http.Request, username string, password string) (string, string, error) {
	err := r.ParseForm()
	if err != nil {
		return "", "", fmt.Errorf("parse form error: %s", err.Error())
	}
	username = r.FormValue("username")
	password = r.FormValue("password")
	return username, password, nil
}

func (s signInController) tokenError(w http.ResponseWriter, err error) model.Error {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusInternalServerError
	w.WriteHeader(http.StatusInternalServerError)
	return e
}

func (s signInController) invalidPasswordError(w http.ResponseWriter, err error) model.Error {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusUnauthorized
	w.WriteHeader(http.StatusUnauthorized)
	return e
}

func (s signInController) usernameNotFoundError(w http.ResponseWriter) model.Error {
	var e model.Error
	e.Message = "username not found"
	e.StatusCode = http.StatusUnauthorized
	w.WriteHeader(http.StatusUnauthorized)
	return e
}

func (s signInController) loadAuthorities(userId int) ([]string, error) {
	return s.userAuthorityRepository.FindAuthoritiesByUserID(userId)
}
