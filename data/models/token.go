package models

import "time"

type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int16     `json:"expires_in,omitempty"`
	IssuedAt     time.Time `json:"issued_at"`
}
