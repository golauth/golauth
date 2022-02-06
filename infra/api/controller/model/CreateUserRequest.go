package model

import (
	"github.com/golauth/golauth/domain/entity"
)

type CreateUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Document  string `json:"document"`
	Password  string `json:"password,omitempty"`
	Enabled   bool   `json:"enabled"`
}

func (u CreateUserRequest) ToEntity() *entity.User {
	return &entity.User{
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Document:  u.Document,
		Password:  u.Password,
		Enabled:   u.Enabled,
	}
}
