package model

import "github.com/dgrijalva/jwt-go"

type Claims struct {
	Username    string   `json:"username"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Authorities []string `json:"authorities,omitempty"`
	jwt.StandardClaims
}
