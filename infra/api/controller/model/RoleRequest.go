package model

import (
	"github.com/golauth/golauth/domain/entity"
	"github.com/google/uuid"
)

type RoleRequest struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func (r RoleRequest) ToEntity() *entity.Role {
	return &entity.Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}
}
