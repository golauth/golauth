package token

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewOpaqueToken(t *testing.T) {
	a, ha, err := newOpaqueToken()
	require.NoError(t, err)
	b, hb, err := newOpaqueToken()
	require.NoError(t, err)

	require.NotEqual(t, a, b, "two calls must not collide")
	require.NotEqual(t, ha, hb)

	raw, err := base64.RawURLEncoding.DecodeString(a)
	require.NoError(t, err)
	require.Len(t, raw, refreshTokenBytes)

	// The hash is a stable, hex SHA-256 of the plaintext.
	require.Len(t, ha, 64)
	require.Equal(t, ha, hashOpaqueToken(a))
}
