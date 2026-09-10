package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// refreshTokenBytes is the entropy of an opaque refresh token. 32 bytes is well
// beyond brute-force reach and keeps the base64url string compact.
const refreshTokenBytes = 32

// newOpaqueToken returns a fresh URL-safe random token and its SHA-256 hex
// digest. The plaintext goes to the client once; only the digest is persisted,
// so a database dump yields no usable session.
func newOpaqueToken() (plain string, hash string, err error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err = rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("could not read random bytes: %w", err)
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	return plain, hashOpaqueToken(plain), nil
}

// hashOpaqueToken is the one-way mapping from a presented token to its stored
// form. A plain SHA-256 is enough here: the input already has 256 bits of
// entropy, so there is nothing for a slow KDF to protect.
func hashOpaqueToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
