package model

type Error struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}
