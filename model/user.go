package model

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Enabled      bool      `json:"enabled"`
	CreationDate time.Time `json:"creationDate"`
}
