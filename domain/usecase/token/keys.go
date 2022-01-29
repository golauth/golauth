package token

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/cristalhq/jwt/v3"
)

func generateSigner(key *rsa.PrivateKey) jwt.Signer {
	signer, err := jwt.NewSignerRS(keyAlgorithm, key)
	if err != nil {
		panic(errSignerGenerate)
	}
	return signer
}

func GeneratePrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Errorf("could not generate private key: %w", err))
	}
	return privateKey
}

func generateVerifier(key *rsa.PrivateKey) jwt.Verifier {
	verifier, err := jwt.NewVerifierRS(keyAlgorithm, &key.PublicKey)
	if err != nil {
		panic(errVerifierGenerate)
	}
	return verifier
}
