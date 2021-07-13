package entity

import (
	"github.com/google/uuid"
	"time"
)

type Role struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Enabled      bool
	CreationDate time.Time
}
