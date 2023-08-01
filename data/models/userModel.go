package models

type User struct {
	Id          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Country     string  `json:"country"`
	SpotifyPlan string  `json:"spotify_plan"`
	Auth        Session `json:"auth"`
}
