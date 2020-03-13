package controller

import (
	"encoding/json"
	"golauth/model"
	"golauth/repository"
	"golauth/util"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type SignupController struct{}

func (s SignupController) CreateUser(w http.ResponseWriter, r *http.Request) {
	userRespository := repository.UserRepository{}
	var user model.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}

	user.Password = string(hash)
	data, err := userRespository.Create(user)
	util.SendResult(w, data, err)
}
