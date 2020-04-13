package model

import "time"

type UserRole struct {
	UserID       int       `json:"userId"`
	RoleID       int       `json:"roleId"`
	CreationDate time.Time `json:"creationDate"`
}
