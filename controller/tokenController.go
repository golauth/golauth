package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"golauth/model"
	"golauth/repository"
	"golauth/util"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type SignInController struct {
	userRepository          repository.UserRepository
	userAuthorityRepository repository.UserAuthorityRepository
}

func NewSignInController(db *sql.DB) SignInController {
	return SignInController{
		userRepository:          repository.NewUserRepository(db),
		userAuthorityRepository: repository.NewUserAuthorityRepository(db),
	}
}

const tokenExpirationTime = 30

func (s SignInController) Token(w http.ResponseWriter, r *http.Request) {

	var username string
	var password string

	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		username, password = s.extractUserPassForm(r, username, password)
	} else if r.Header.Get("Content-Type") == "application/json" {
		username, password = s.extractUserPassJson(r, username, password)
	} else {
		util.SendBadRequest(w, errors.New("Content-Type not supported"))
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

	jwtToken, err := s.generateJwtToken(user, authorities)
	if err != nil {
		e := s.tokenError(w, err)
		_ = json.NewEncoder(w).Encode(e)
		return
	}
	tokenResponse := model.TokenResponse{AccessToken: jwtToken}
	util.SendSuccess(w, tokenResponse)
}

func (s SignInController) extractUserPassJson(r *http.Request, username string, password string) (string, string) {
	var userLogin model.UserLogin
	_ = json.NewDecoder(r.Body).Decode(&userLogin)
	username = userLogin.Username
	password = userLogin.Password
	return username, password
}

func (s SignInController) extractUserPassForm(r *http.Request, username string, password string) (string, string) {
	err := r.ParseForm()
	util.LogError(err)
	username = r.FormValue("username")
	password = r.FormValue("password")
	return username, password
}

func (s SignInController) tokenError(w http.ResponseWriter, err error) model.Error {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusInternalServerError
	w.WriteHeader(http.StatusInternalServerError)
	return e
}

func (s SignInController) invalidPasswordError(w http.ResponseWriter, err error) model.Error {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusUnauthorized
	w.WriteHeader(http.StatusUnauthorized)
	return e
}

func (s SignInController) usernameNotFoundError(w http.ResponseWriter) model.Error {
	var e model.Error
	e.Message = "username not found"
	e.StatusCode = http.StatusUnauthorized
	w.WriteHeader(http.StatusUnauthorized)
	return e
}

func (s SignInController) loadAuthorities(userId int) ([]string, error) {
	return s.userAuthorityRepository.FindAuthoritiesByUserID(userId)
}

func (s SignInController) generateJwtToken(user model.User, authorities []string) (interface{}, error) {
	expirationTime := time.Now().Add(tokenExpirationTime * time.Minute)
	claims := &model.Claims{
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Authorities: authorities,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(util.SignKey)
}
