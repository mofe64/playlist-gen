package models

type Item struct {
	Explicit   bool     `json:"explicit,omitempty"`
	Href       string   `json:"href,omitempty"`
	Id         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	Popularity int16    `json:"popularity,omitempty"`
	Uri        string   `json:"uri,omitempty"`
	Genres     []string `json:"genres,omitempty"`
	Type       string   `json:"type,omitempty"`
	Artists    []Artist `json:"artists,omitempty"`
	Album      Album    `json:"album,omitempty"`
}
