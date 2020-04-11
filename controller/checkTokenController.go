package controller

import (
	"golauth/model"
	"golauth/util"
	"net/http"
)

type CheckTokenController struct{}

func (c CheckTokenController) CheckToken(w http.ResponseWriter, r *http.Request) {
	token, err := util.ExtractToken(r)
	if err == (model.Error{}) {
		util.SendError(w, err)
		return
	}
	err = util.ValidateToken(token)
	if err == (model.Error{}) {
		util.SendError(w, err)
		return
	}
	util.SendSuccess(w, true)
}
