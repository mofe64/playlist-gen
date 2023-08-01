package responses

import "mofe64/playlistGen/data/models"

type TopItemsResponse struct {
	Href     string         `json:"href"`
	Limit    int16          `json:"limit"`
	Next     string         `json:"next"`
	Offset   int16          `json:"offset"`
	Previous string         `json:"previous"`
	Total    int16          `json:"total"`
	Items    []models.Track `json:"items"`
}
