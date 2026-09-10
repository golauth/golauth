// Package keys is the single seam through which golauth obtains its JWT signing
// material. Load reads the key from configuration so that it is identical across
// replicas and survives restarts; every key carries a deterministic id and its
// public half is published through KeySet.JWKS.
package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Environment variables consulted by Load.
const (
	EnvPrivateKey         = "JWT_PRIVATE_KEY"          // PEM content, inline
	EnvPrivateKeyFile     = "JWT_PRIVATE_KEY_FILE"     // path to a PEM file
	EnvPrivateKeyPrevious = "JWT_PRIVATE_KEY_PREVIOUS" // verify-only PEM blocks
	EnvAppEnv             = "APP_ENV"                  // "production" fails closed
)

// minRSABits is the smallest modulus accepted for an RS512 signing key.
const minRSABits = 2048

var (
	// ErrNoSigningKey is returned by Load when APP_ENV=production and no key
	// was supplied. Load never exits the process on its own; main turns this
	// into a fatal so the loader stays testable.
	ErrNoSigningKey = errors.New("keys: no signing key configured and APP_ENV=production")

	errKeyTooSmall = fmt.Errorf("keys: RSA signing key must be at least %d bits", minRSABits)
	errNoPEMBlock  = errors.New("keys: no PEM block found in key material")
	errNotRSAKey   = errors.New("keys: PEM block does not contain an RSA private key")
)

// SigningKey is a parsed RSA private key together with its derived key id.
type SigningKey struct {
	Private *rsa.PrivateKey
	KID     string
}

// KeySet is the material a golauth instance trusts: exactly one Current key that
// signs new tokens, plus zero or more Previous keys kept only to verify tokens
// minted before the last rotation. Publishing every public half is what turns a
// rotation into a rolling restart instead of a mass logout.
type KeySet struct {
	Current  *SigningKey
	Previous []*SigningKey
}

// All returns the current key followed by every previous key, in a fresh slice.
func (ks *KeySet) All() []*SigningKey {
	out := make([]*SigningKey, 0, 1+len(ks.Previous))
	out = append(out, ks.Current)
	out = append(out, ks.Previous...)
	return out
}

// Load builds the KeySet from the environment. Resolution order for the signing
// key:
//
//  1. JWT_PRIVATE_KEY               -- PEM content inline
//  2. JWT_PRIVATE_KEY_FILE          -- path to a mounted PEM file
//  3. nothing set, APP_ENV unset/"dev" -- an ephemeral key, with a loud warning
//  4. nothing set, APP_ENV=production   -- ErrNoSigningKey
//
// JWT_PRIVATE_KEY_PREVIOUS carries keys that no longer sign but must still
// verify; it accepts one or more concatenated PEM blocks.
func Load() (*KeySet, error) {
	current, err := loadCurrent()
	if err != nil {
		return nil, err
	}

	previous, err := parseSigningKeys([]byte(os.Getenv(EnvPrivateKeyPrevious)))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", EnvPrivateKeyPrevious, err)
	}

	return &KeySet{Current: current, Previous: withoutKID(previous, current.KID)}, nil
}

// Generate returns a KeySet backed by a fresh ephemeral key. It is used for
// local development and by tests; deployments go through Load.
func Generate() *KeySet {
	priv, err := rsa.GenerateKey(rand.Reader, minRSABits)
	if err != nil {
		panic(fmt.Errorf("keys: generate ephemeral key: %w", err))
	}
	sk, err := newSigningKey(priv)
	if err != nil {
		panic(err)
	}
	return &KeySet{Current: sk}
}

func loadCurrent() (*SigningKey, error) {
	if inline := strings.TrimSpace(os.Getenv(EnvPrivateKey)); inline != "" {
		sk, err := parseOne([]byte(inline))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvPrivateKey, err)
		}
		return sk, nil
	}

	if path := strings.TrimSpace(os.Getenv(EnvPrivateKeyFile)); path != "" {
		// #nosec G304 G703 -- the path is an operator-supplied secret mount
		// point (a Kubernetes secret file), not attacker-controlled input.
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvPrivateKeyFile, err)
		}
		sk, err := parseOne(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvPrivateKeyFile, err)
		}
		return sk, nil
	}

	if isProduction() {
		return nil, ErrNoSigningKey
	}

	slog.Warn("keys: no signing key configured; using an ephemeral key. "+
		"Every restart invalidates all outstanding tokens and a second replica cannot verify them. "+
		"Set one of these variables before deploying.",
		"vars", EnvPrivateKey+", "+EnvPrivateKeyFile)
	return Generate().Current, nil
}

func isProduction() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(EnvAppEnv)), "production")
}

// parseOne decodes the first PEM block of pemBytes into a signing key.
func parseOne(pemBytes []byte) (*SigningKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errNoPEMBlock
	}
	priv, err := parseBlock(block)
	if err != nil {
		return nil, err
	}
	return newSigningKey(priv)
}

// parseSigningKeys decodes every PEM block found in blob. An empty blob yields
// no keys and no error.
func parseSigningKeys(blob []byte) ([]*SigningKey, error) {
	var keys []*SigningKey
	rest := blob
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return keys, nil
		}
		priv, err := parseBlock(block)
		if err != nil {
			return nil, err
		}
		sk, err := newSigningKey(priv)
		if err != nil {
			return nil, err
		}
		keys = append(keys, sk)
	}
}

// parseBlock accepts either a PKCS#1 ("RSA PRIVATE KEY") or a PKCS#8
// ("PRIVATE KEY") RSA key.
func parseBlock(block *pem.Block) (*rsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("keys: parse private key: %w", err)
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errNotRSAKey
	}
	return rsaKey, nil
}

// newSigningKey validates the modulus size and derives the key id.
func newSigningKey(priv *rsa.PrivateKey) (*SigningKey, error) {
	if priv.N.BitLen() < minRSABits {
		return nil, errKeyTooSmall
	}
	kid, err := deriveKID(&priv.PublicKey)
	if err != nil {
		return nil, err
	}
	return &SigningKey{Private: priv, KID: kid}, nil
}

// deriveKID is the base64url-encoded SHA-256 of the DER (PKIX) encoding of the
// public key. It is deterministic, so two replicas loading the same PEM publish
// the same id with no shared configuration.
func deriveKID(pub *rsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("keys: marshal public key: %w", err)
	}
	sum := sha256.Sum256(der)
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

// withoutKID drops any key whose id equals kid, so a key listed in both the
// current and the previous slot is not verified (or published) twice.
func withoutKID(keys []*SigningKey, kid string) []*SigningKey {
	if len(keys) == 0 {
		return nil
	}
	out := make([]*SigningKey, 0, len(keys))
	for _, sk := range keys {
		if sk.KID != kid {
			out = append(out, sk)
		}
	}
	return out
}
