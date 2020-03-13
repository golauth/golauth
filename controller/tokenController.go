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

	err := r.ParseForm()
	util.LogFatal(err)

	username := r.FormValue("username")
	password := r.FormValue("password")

	data, err := userRespository.FindByUsername(username)

	if (model.User{}) == data {
		util.SendNotFound(w, errors.New("username not found"))
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

	token, err := generateJwtToken(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	util.SendSuccess(w, token)
}

func comparePassword(pwPlain string, dbPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dbPw), []byte(pwPlain))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func generateJwtToken(user model.User) (interface{}, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(signKey)
}
