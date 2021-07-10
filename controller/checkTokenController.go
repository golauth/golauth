//go:generate mockgen -source checkTokenController.go -destination mock/checkTokenController_mock.go -package mock
package controller

import (
	"encoding/json"
	"golauth/model"
	"golauth/usecase"
	"net/http"
)

type CheckTokenController interface {
	CheckToken(w http.ResponseWriter, r *http.Request)
}

type checkTokenController struct {
	service usecase.TokenService
}

func NewCheckTokenController(service usecase.TokenService) CheckTokenController {
	return checkTokenController{service: service}
}

func (c checkTokenController) CheckToken(w http.ResponseWriter, r *http.Request) {
	token, err := c.service.ExtractToken(r)
	if err != nil {
		sendBadRequest(w, err)
		return
	}
	err = c.service.ValidateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(&model.Error{StatusCode: http.StatusUnauthorized, Message: err.Error()})
		return
	}
	sendSuccess(w, true)
}
