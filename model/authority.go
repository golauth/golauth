package model

import "time"

type Authority struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Enabled      bool      `json:"enabled"`
	CreationDate time.Time `json:"creationDate"`
}
