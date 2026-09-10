package keys

import (
	"encoding/base64"
	"math/big"
)

// JWK is a single RSA public key in JSON Web Key form (RFC 7517).
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS is a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKS renders every trusted public key (current first) as a key set. The
// document carries no private material and is safe to serve publicly.
func (ks *KeySet) JWKS() JWKS {
	all := ks.All()
	out := JWKS{Keys: make([]JWK, 0, len(all))}
	for _, sk := range all {
		out.Keys = append(out.Keys, publicJWK(sk))
	}
	return out
}

func publicJWK(sk *SigningKey) JWK {
	pub := &sk.Private.PublicKey
	return JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS512",
		Kid: sk.KID,
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(exponentBytes(pub.E)),
	}
}

// exponentBytes returns the minimal big-endian encoding of a public exponent,
// as required for the JWK "e" parameter (65537 becomes "AQAB").
func exponentBytes(e int) []byte {
	return new(big.Int).SetInt64(int64(e)).Bytes()
}
