package util

import (
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/subosito/gotenv"
	"golauth/model"
	"io/ioutil"
	"log"
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
	if err != nil {
		log.Fatal(fmt.Errorf("could not read private key path \"PRIVATE_KEY_PATH[%s]\": %w", privateKeyPath, err))
	}

	SignKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(fmt.Errorf("could not parse RSA private key from pem: %w", err))
	}

	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal(fmt.Errorf("could not read public key path \"PUBLIC_KEY_PATH[%s]\": %w", privateKeyPath, err))
	}

	VerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(fmt.Errorf("could not parse RSA public key from pem: %w", err))
	}
}

func ValidateToken(token string) *model.Error {
	claims := &model.Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (i interface{}, err error) {
		return VerifyKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return &model.Error{Message: err.Error(), StatusCode: http.StatusUnauthorized}
		}
		return &model.Error{Message: err.Error(), StatusCode: http.StatusUnauthorized}
	}

	err = parsedToken.Claims.Valid()
	if err != nil {
		return &model.Error{Message: err.Error(), StatusCode: http.StatusUnauthorized}
	}

	if !parsedToken.Valid {
		return &model.Error{Message: "Invalid token", StatusCode: http.StatusUnauthorized}
	}

	return nil
}

func ExtractToken(r *http.Request) (string, *model.Error) {
	authorization := r.Header.Get("Authorization")
	if len(authorization) > len("Bearer ") {
		return authorization[7:], nil
	}
	return "", &model.Error{Message: "bearer token extract error", StatusCode: http.StatusBadGateway}
}
