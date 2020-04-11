package util

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/subosito/gotenv"
	"golauth/model"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	privateKeyPath string
	publicKeyPath  string
	VerifyKey      *rsa.PublicKey
	SignKey        *rsa.PrivateKey
)

func init() {
	_ = gotenv.Load()

	privateKeyPath = os.Getenv("PRIVATE_KEY_PATH")
	publicKeyPath = os.Getenv("PUBLIC_KEY_PATH")

	signBytes, err := ioutil.ReadFile(privateKeyPath)
	LogFatal(err)

	SignKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	LogFatal(err)

	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	LogFatal(err)

	VerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	LogFatal(err)
}

func ValidateToken(token string) model.Error {
	claims := &model.Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (i interface{}, err error) {
		return VerifyKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return model.Error{Message: err.Error(), StatusCode: http.StatusUnauthorized}
		}
		return model.Error{Message: err.Error(), StatusCode: http.StatusUnauthorized}
	}

	err = parsedToken.Claims.Valid()
	if err != nil {
		return model.Error{Message: err.Error(), StatusCode: http.StatusUnauthorized}
	}

	if !parsedToken.Valid {
		return model.Error{Message: "Invalid token", StatusCode: http.StatusUnauthorized}
	}

	return model.Error{}
}

func ExtractToken(r *http.Request) (string, model.Error) {
	authorization := r.Header.Get("Authorization")
	if len(authorization) > len("Bearer ") {
		return authorization[7:], model.Error{}
	}
	return "", model.Error{Message: "bearer token extract error"}
}
