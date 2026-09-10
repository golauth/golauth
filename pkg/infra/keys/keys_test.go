package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func pkcs1PEM(t *testing.T, bits int) (string, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, bits)
	require.NoError(t, err)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return string(pem.EncodeToMemory(block)), key
}

func pkcs8PEM(t *testing.T, bits int) (string, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, bits)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), key
}

// clearEnv makes sure a stray value from the real environment does not leak
// into a test that expects a variable to be unset.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{EnvPrivateKey, EnvPrivateKeyFile, EnvPrivateKeyPrevious, EnvAppEnv} {
		t.Setenv(k, "")
	}
}

func TestLoadInlinePKCS1(t *testing.T) {
	clearEnv(t)
	pemStr, key := pkcs1PEM(t, 2048)
	t.Setenv(EnvPrivateKey, pemStr)

	ks, err := Load()
	require.NoError(t, err)
	require.Equal(t, key.N, ks.Current.Private.N)
	require.NotEmpty(t, ks.Current.KID)
	require.Empty(t, ks.Previous)
}

func TestLoadInlinePKCS8(t *testing.T) {
	clearEnv(t)
	pemStr, key := pkcs8PEM(t, 2048)
	t.Setenv(EnvPrivateKey, pemStr)

	ks, err := Load()
	require.NoError(t, err)
	require.Equal(t, key.N, ks.Current.Private.N)
}

func TestLoadFromFile(t *testing.T) {
	clearEnv(t)
	pemStr, key := pkcs8PEM(t, 2048)
	path := t.TempDir() + "/jwt-key.pem"
	require.NoError(t, os.WriteFile(path, []byte(pemStr), 0o600))
	t.Setenv(EnvPrivateKeyFile, path)

	ks, err := Load()
	require.NoError(t, err)
	require.Equal(t, key.N, ks.Current.Private.N)
}

func TestLoadInlineWinsOverFile(t *testing.T) {
	clearEnv(t)
	inline, inlineKey := pkcs8PEM(t, 2048)
	filePem, _ := pkcs8PEM(t, 2048)
	path := t.TempDir() + "/jwt-key.pem"
	require.NoError(t, os.WriteFile(path, []byte(filePem), 0o600))
	t.Setenv(EnvPrivateKey, inline)
	t.Setenv(EnvPrivateKeyFile, path)

	ks, err := Load()
	require.NoError(t, err)
	require.Equal(t, inlineKey.N, ks.Current.Private.N)
}

func TestLoadMalformedPEM(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvPrivateKey, "-----BEGIN RSA PRIVATE KEY-----\nnot base64\n-----END RSA PRIVATE KEY-----")

	_, err := Load()
	require.Error(t, err)
}

func TestLoadNoPEMBlock(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvPrivateKey, "just some text")

	_, err := Load()
	require.ErrorIs(t, err, errNoPEMBlock)
}

func TestLoadKeyUnder2048(t *testing.T) {
	clearEnv(t)
	pemStr, _ := pkcs1PEM(t, 1024)
	t.Setenv(EnvPrivateKey, pemStr)

	_, err := Load()
	require.ErrorIs(t, err, errKeyTooSmall)
}

func TestLoadProductionWithoutKeyFails(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvAppEnv, "production")

	_, err := Load()
	require.ErrorIs(t, err, ErrNoSigningKey)
}

func TestLoadDevWithoutKeyGeneratesEphemeral(t *testing.T) {
	clearEnv(t)

	ks, err := Load()
	require.NoError(t, err)
	require.NotNil(t, ks.Current)
	require.GreaterOrEqual(t, ks.Current.Private.N.BitLen(), minRSABits)
	require.NotEmpty(t, ks.Current.KID)
}

func TestKIDIsDeterministic(t *testing.T) {
	pemStr, _ := pkcs8PEM(t, 2048)

	clearEnv(t)
	t.Setenv(EnvPrivateKey, pemStr)
	first, err := Load()
	require.NoError(t, err)

	second, err := Load()
	require.NoError(t, err)
	require.Equal(t, first.Current.KID, second.Current.KID)
}

func TestLoadPreviousKeysParsedAndDeduped(t *testing.T) {
	clearEnv(t)
	current, _ := pkcs8PEM(t, 2048)
	retired, _ := pkcs8PEM(t, 2048)
	t.Setenv(EnvPrivateKey, current)
	// The previous list carries the retired key and, redundantly, the current
	// one: the current must not be verified or published twice.
	t.Setenv(EnvPrivateKeyPrevious, retired+"\n"+current)

	ks, err := Load()
	require.NoError(t, err)
	require.Len(t, ks.Previous, 1)
	require.NotEqual(t, ks.Current.KID, ks.Previous[0].KID)

	all := ks.All()
	require.Len(t, all, 2)
}

func TestJWKSRendersEveryTrustedKey(t *testing.T) {
	clearEnv(t)
	current, _ := pkcs8PEM(t, 2048)
	retired, _ := pkcs8PEM(t, 2048)
	t.Setenv(EnvPrivateKey, current)
	t.Setenv(EnvPrivateKeyPrevious, retired)

	ks, err := Load()
	require.NoError(t, err)

	doc := ks.JWKS()
	require.Len(t, doc.Keys, 2)
	require.Equal(t, ks.Current.KID, doc.Keys[0].Kid)
	for _, k := range doc.Keys {
		require.Equal(t, "RSA", k.Kty)
		require.Equal(t, "sig", k.Use)
		require.Equal(t, "RS512", k.Alg)
		require.Equal(t, "AQAB", k.E)
		require.NotEmpty(t, k.N)
	}
}
