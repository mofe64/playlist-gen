package responses

import "mofe64/playlistGen/data/models"

type PlaylistCreationResponse struct {
	Name  string      `json:"name,omitempty"`
	Owner models.User `json:"owner"`
}
