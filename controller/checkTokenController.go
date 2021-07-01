package controller

import (
	"golauth/model"
	"golauth/util"
	"net/http"
)

type CheckTokenController struct{}

func NewCheckTokenController() CheckTokenController {
	return CheckTokenController{}
}

func (c CheckTokenController) CheckToken(w http.ResponseWriter, r *http.Request) {
	token, err := util.ExtractToken(r)
	if err != nil {
		util.SendError(w, &model.Error{StatusCode: http.StatusBadGateway, Message: err.Error()})
		return
	}
	err = util.ValidateToken(token)
	if err != nil {
		util.SendError(w, &model.Error{StatusCode: http.StatusUnauthorized, Message: err.Error()})
		return
	}
	util.SendSuccess(w, true)
}
