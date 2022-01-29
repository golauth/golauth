package model

import (
	"github.com/google/uuid"
)

type UserRoleRequest struct {
	UserID uuid.UUID `json:"userId"`
	RoleID uuid.UUID `json:"roleId"`
}
