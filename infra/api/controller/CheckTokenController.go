package controller

import (
	"encoding/json"
	"github.com/golauth/golauth/core/util"
	"github.com/golauth/golauth/domain/usecase/token"
	"net/http"
)

type CheckTokenController interface {
	CheckToken(w http.ResponseWriter, r *http.Request)
}

type checkTokenController struct {
	validateToken token.ValidateToken
}

func NewCheckTokenController(validateToken token.ValidateToken) CheckTokenController {
	return checkTokenController{validateToken: validateToken}
}

func (c checkTokenController) CheckToken(w http.ResponseWriter, r *http.Request) {
	t, err := util.ExtractToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = c.validateToken.Execute(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(true)
}
