package models

type Track struct {
	Explicit   bool   `json:"explicit"`
	Href       string `json:"href"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	Popularity int16  `json:"popularity"`
	Uri        string `json:"uri"`
}
