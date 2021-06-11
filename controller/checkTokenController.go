package controller

import (
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
		util.SendError(w, err)
		return
	}
	err = util.ValidateToken(token)
	if err != nil {
		util.SendError(w, err)
		return
	}
	util.SendSuccess(w, true)
}
