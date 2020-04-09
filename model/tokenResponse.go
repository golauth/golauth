package model

type TokenResponse struct {
	AccessToken  interface{} `json:"access_token"`
	RefreshToken interface{} `json:"refresh_token"`
}
