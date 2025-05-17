package entity

import (
	"github.com/google/uuid"
	"time"
)

type RoleAuthority struct {
	RoleID       uuid.UUID
	AuthorityID  string
	CreationDate time.Time
}
