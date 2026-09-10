package entity

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken is one long-lived, opaque session credential. The plaintext is
// returned to the client exactly once at issue time; only TokenHash (a SHA-256
// hex digest) is persisted.
type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	IssuedAt   time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *uuid.UUID
	UserAgent  string
	ClientIP   string
}

// Revoked reports whether the token has been explicitly invalidated (logout,
// family revocation, or rotation).
func (t RefreshToken) Revoked() bool { return t.RevokedAt != nil }

// Expired reports whether the token is past its lifetime at the given instant.
func (t RefreshToken) Expired(now time.Time) bool { return !now.Before(t.ExpiresAt) }

// Rotated reports whether a newer token has already superseded this one. A
// rotated token presented again is reuse and must burn the whole family.
func (t RefreshToken) Rotated() bool { return t.ReplacedBy != nil }

// Usable reports whether the token may still be exchanged for a new pair.
func (t RefreshToken) Usable(now time.Time) bool {
	return !t.Revoked() && !t.Rotated() && !t.Expired(now)
}
