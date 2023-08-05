package models

type Playlist struct {
	Name        string `json:"name,omitempty"`
	Owner       User   `json:"owner"`
	Description string `json:"description"`
	Href        string `json:"href,omitempty"`
	Id          string `json:"id,omitempty"`
	Public      bool   `json:"public"`
	SnapshotId  string `json:"snapshot_id"`
	Tracks      Tracks `json:"tracks"`
	URI         string `json:"uri"`
}

type Tracks struct {
	Href     string          `json:"href,omitempty"`
	Limit    int16           `json:"limit"`
	Next     string          `json:"next"`
	Offset   int16           `json:"offset"`
	Previous string          `json:"previous"`
	Total    int16           `json:"total"`
	Items    []PlaylistTrack `json:"items"`
}

type PlaylistTrack struct {
	AddedAt string `json:"addedd_at"`
	Track   Track  `json:"track"`
}
