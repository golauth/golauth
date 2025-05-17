package model

import (
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/google/uuid"
	"time"
)

type RoleResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Enabled      bool      `json:"enabled"`
	CreationDate time.Time `json:"creationDate"`
}

func NewRoleResponseFromEntity(e *entity.Role) *RoleResponse {
	return &RoleResponse{
		ID:           e.ID,
		Name:         e.Name,
		Description:  e.Description,
		Enabled:      e.Enabled,
		CreationDate: e.CreationDate,
	}
}
