package util

import (
	"fmt"
	"github.com/subosito/gotenv"
	"io/ioutil"
	"os"
)

func LoadKeyFromEnv() ([]byte, []byte, error) {
	_ = gotenv.Load()

	privateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")

	privBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read private key path \"PRIVATE_KEY_PATH[%s]\": %w", privateKeyPath, err)
	}

	pubBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read public key path \"PUBLIC_KEY_PATH[%s]\": %w", privateKeyPath, err)
	}

	return privBytes, pubBytes, nil
}
