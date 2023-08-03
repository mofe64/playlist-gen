package responses

type TracksAudioFeatures struct {
	AudioFeatures []Features `json:"audio_features"`
}

type Features struct {
	Acousticness     float32 `json:"acousticness"`
	Danceability     float32 `json:"danceability"`
	DurationMs       int32   `json:"duration_ms"`
	Energy           float32 `json:"energy"`
	Id               string  `json:"id"`
	Instrumentalness float32 `json:"instrumentalness"`
	Liveness         float32 `json:"liveness"`
	Tempo            float64 `json:"tempo"`
	Valence          float32 `json:"valence"`
}
