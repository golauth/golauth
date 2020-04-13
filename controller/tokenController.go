package controller

import (
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

type SigninController struct{}

func (s SigninController) Token(w http.ResponseWriter, r *http.Request) {
	userRespository := repository.UserRepository{}

	var username string
	var password string

	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		err := r.ParseForm()
		util.LogError(err)

		username = r.FormValue("username")
		password = r.FormValue("password")
	} else if r.Header.Get("Content-Type") == "application/json" {
		var userLogin model.UserLogin
		_ = json.NewDecoder(r.Body).Decode(&userLogin)
		username = userLogin.Username
		password = userLogin.Password
	} else {
		util.SendBadRequest(w, errors.New("Content-Type not supported"))
	}

	data, err := userRespository.FindByUsernameWithPassword(username)
	if (model.User{}) == data {
		var e model.Error
		e.Message = "username not found"
		e.StatusCode = http.StatusUnauthorized
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(e)
		return
	}

	user := data.(model.User)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		var e model.Error
		e.Message = err.Error()
		e.StatusCode = http.StatusUnauthorized
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(e)
		return
	}

	authorities, _ := loadAuthorities(user.ID)

	jwtToken, err := generateJwtToken(user, authorities)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenResponse := model.TokenResponse{AccessToken: jwtToken}
	util.SendSuccess(w, tokenResponse)
}

func loadAuthorities(userId int) ([]string, error) {
	userAuthorityRepository := repository.UserAuthorityRepository{}
	return userAuthorityRepository.FindAuthoritiesByUserID(userId)
}

func generateJwtToken(user model.User, authorities []string) (interface{}, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
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
