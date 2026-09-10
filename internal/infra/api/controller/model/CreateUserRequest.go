package model

import (
	"github.com/golauth/golauth/internal/domain/entity"
)

// CreateUserRequest is the public signup payload. It deliberately has no
// "enabled" field: account activation is an administrative operation, and the
// old field was decoded and then silently overwritten -- a misleading contract.
type CreateUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Document  string `json:"document"`
	Password  string `json:"password,omitempty"`
}

func (u CreateUserRequest) ToEntity() *entity.User {
	return &entity.User{
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Document:  u.Document,
		Password:  u.Password,
	}
}
