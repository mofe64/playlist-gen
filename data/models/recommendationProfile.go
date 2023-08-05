package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type RecommendationProfile struct {
	Id               primitive.ObjectID `json:"id" bson:"_id"`
	CreatorId        string             `json:"creator_id"`
	PlaylistName     string             `json:"playlist_name"`
	Limit            int16              `json:"limit"`
	SeedArtists      []string           `json:"seed_artists"`
	SeedGenres       []string           `json:"seed_genres"`
	SeedTracks       []string           `json:"seed_tracks"`
	Acousticness     float32            `json:"acousticness"`
	Danceability     float32            `json:"danceability"`
	Energy           float32            `json:"energy"`
	Instrumentalness float32            `json:"instrumentalness"`
	Liveness         float32            `json:"liveness"`
	Popularity       int16              `json:"popularity"`
	Valence          float32            `json:"valence"`
	Tempo            float32            `json:"tempo"`
	SnapshotId       string             `json:"snapshot_id"`
}
