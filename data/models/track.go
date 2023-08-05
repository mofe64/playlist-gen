package models

type Track struct {
	Explicit   bool     `json:"explicit,omitempty"`
	Href       string   `json:"href,omitempty"`
	Id         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	Popularity int16    `json:"popularity,omitempty"`
	Uri        string   `json:"uri,omitempty"`
	Artists    []Artist `json:"artists,omitempty"`
}

type Artist struct {
	Id         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	Popularity int16    `json:"popularity,omitempty"`
	Genres     []string `json:"genres,omitempty"`
}
type Album struct {
	AlbumType  string   `json:"album_type,omitempty"`
	Href       string   `json:"href,omitempty"`
	Id         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	Uri        string   `json:"uri,omitempty"`
	Genres     []string `json:"genres,omitempty"`
	Popularity int16    `json:"popularity,omitempty"`
}
