package controller

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"golauth/model"
	"golauth/util"
	"net/http"
)

type CheckTokenController struct{}

func (c CheckTokenController) CheckToken(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	token := authorization[7:]

	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (i interface{}, err error) {
		return verifyKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			sendError(w, http.StatusUnauthorized, err.Error())
			return
		}
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !parsedToken.Valid {
		sendError(w, http.StatusUnauthorized, "Invalid token.")
		return
	}

	util.SendSuccess(w, true)
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	var e model.Error
	e.Message = message
	e.StatusCode = statusCode
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(e)
}
