package model

import (
	"github.com/google/uuid"
	"time"
)

type UserRequest struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Document  string    `json:"document"`
	Password  string    `json:"password,omitempty"`
	Enabled   bool      `json:"enabled"`
}

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Document     string    `json:"document"`
	Enabled      bool      `json:"enabled"`
	CreationDate time.Time `json:"creationDate"`
}

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRoleRequest struct {
	UserID uuid.UUID `json:"userId"`
	RoleID uuid.UUID `json:"roleId"`
}
