package model

import (
	"github.com/cristalhq/jwt/v3"
)

type Claims struct {
	Username    string   `json:"username"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Authorities []string `json:"authorities,omitempty"`
	jwt.StandardClaims
}
