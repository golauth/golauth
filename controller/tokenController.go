package controller

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"golauth/model"
	"golauth/repository"
	"golauth/util"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/subosito/gotenv"
	"golang.org/x/crypto/bcrypt"
)

type SigninController struct{}

type Claims struct {
	Username    string   `json:"username"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Authorities []string `json:"authorities,omitempty"`
	jwt.StandardClaims
}

var (
	privateKeyPath string
	publicKeyPath  string
	verifyKey      *rsa.PublicKey
	signKey        *rsa.PrivateKey
)

func init() {
	_ = gotenv.Load()

	privateKeyPath = os.Getenv("PRIVATE_KEY_PATH")
	publicKeyPath = os.Getenv("PUBLIC_KEY_PATH")

	signBytes, err := ioutil.ReadFile(privateKeyPath)
	util.LogFatal(err)

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	util.LogFatal(err)

	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	util.LogFatal(err)

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	util.LogFatal(err)
}

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

	data, err := userRespository.FindByUsername(username)
	if (model.User{}) == data {
		var e model.Error
		e.Message = "username not found"
		e.StatusCode = http.StatusUnauthorized
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(e)
		return
	}

	user := data.(model.User)
	if !comparePassword(password, user.Password) {
		var e model.Error
		e.Message = "password doesn't match"
		e.StatusCode = http.StatusUnauthorized
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(e)
		return
	}

	authorities, _ := loadAuthorities(user.ID)

	token, err := generateJwtToken(user, authorities)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	util.SendSuccess(w, token)
}

func loadAuthorities(userId int) ([]string, error) {
	userAuthorityRepository := repository.UserAuthorityRepository{}
	return userAuthorityRepository.FindAuthoritiesByUserID(userId)
}

func comparePassword(pwPlain string, dbPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dbPw), []byte(pwPlain))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func generateJwtToken(user model.User, authorities []string) (interface{}, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Authorities: authorities,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(signKey)
}
