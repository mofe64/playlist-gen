package responses

import "mofe64/playlistGen/data/models"

type RecommendationsResponse struct {
	Seeds  []Seed         `json:"seeds"`
	Tracks []models.Track `json:"tracks"`
}

type Seed struct {
	AfterFilteringSize int16  `json:"afterFilteringSize"`
	AfterRelinkingSize int16  `json:"afterRelinkingSize"`
	Href               string `json:"href"`
	Id                 string `json:"id"`
	InitialPoolSize    int16  `json:"initialPoolSize"`
	Type               string `json:"type"`
}
