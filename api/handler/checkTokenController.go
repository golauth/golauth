//go:generate mockgen -source checkTokenController.go -destination mock/checkTokenController_mock.go -package mock
package handler

import (
	"encoding/json"
	"golauth/usecase"
	"net/http"
)

type CheckTokenController interface {
	CheckToken(w http.ResponseWriter, r *http.Request)
}

type checkTokenController struct {
	svc usecase.TokenService
}

func NewCheckTokenController(s usecase.TokenService) CheckTokenController {
	return checkTokenController{svc: s}
}

func (c checkTokenController) CheckToken(w http.ResponseWriter, r *http.Request) {
	token, err := c.svc.ExtractToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = c.svc.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(true)
}
