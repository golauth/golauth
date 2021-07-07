//go:generate mockgen -source checkTokenController.go -destination mock/checkTokenController_mock.go -package mock
package controller

import (
	"golauth/model"
	"golauth/usercase"
	"golauth/util"
	"net/http"
)

type CheckTokenController interface {
	CheckToken(w http.ResponseWriter, r *http.Request)
}

type checkTokenController struct {
	svc usercase.TokenService
}

func NewCheckTokenController(privBytes []byte, pubBytes []byte) CheckTokenController {
	return checkTokenController{svc: usercase.NewTokenService(privBytes, pubBytes)}
}

func (c checkTokenController) CheckToken(w http.ResponseWriter, r *http.Request) {
	token, err := c.svc.ExtractToken(r)
	if err != nil {
		util.SendError(w, &model.Error{StatusCode: http.StatusBadGateway, Message: err.Error()})
		return
	}
	err = c.svc.ValidateToken(token)
	if err != nil {
		util.SendError(w, &model.Error{StatusCode: http.StatusUnauthorized, Message: err.Error()})
		return
	}
	util.SendSuccess(w, true)
}
