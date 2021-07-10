//go:generate mockgen -source checkTokenController.go -destination mock/checkTokenController_mock.go -package mock
package controller

import (
	"golauth/model"
	"golauth/usecase"
	"golauth/util"
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
		util.SendError(w, &model.Error{StatusCode: http.StatusBadGateway, Message: err.Error()})
		return
	}
	err = c.service.ValidateToken(token)
	if err != nil {
		util.SendError(w, &model.Error{StatusCode: http.StatusUnauthorized, Message: err.Error()})
		return
	}
	util.SendSuccess(w, true)
}
