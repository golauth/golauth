package handler

import (
	"encoding/json"
	"github.com/golauth/golauth/domain/usecase/token"
	"net/http"
)

type CheckTokenController interface {
	CheckToken(w http.ResponseWriter, r *http.Request)
}

type checkTokenController struct {
	svc token.UseCase
}

func NewCheckTokenController(s token.UseCase) CheckTokenController {
	return checkTokenController{svc: s}
}

func (c checkTokenController) CheckToken(w http.ResponseWriter, r *http.Request) {
	t, err := c.svc.ExtractToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = c.svc.ValidateToken(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(true)
}
