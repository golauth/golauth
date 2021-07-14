package entity

import (
	"github.com/google/uuid"
	"time"
)

type UserRole struct {
	UserID       uuid.UUID
	RoleID       uuid.UUID
	Enabled      bool
	CreationDate time.Time
}
