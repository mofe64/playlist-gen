package handlers

import (
	"context"
	"encoding/json"
	"mofe64/playlistGen/config"
	"mofe64/playlistGen/data/models"
	"mofe64/playlistGen/data/responses"
	"mofe64/playlistGen/util"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

const CUTOFF = 50

var recommendationProfileCollection = config.GetCollection(config.DATABASE, "recommendationProfiles")

func CreatePlaylist() gin.HandlerFunc {
	tag := "CREATE_PLAYLIST_HANDLER"
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		userId := c.Param("userId")
		value := c.GetString("userDetails")
		util.InfoLog.Println(tag+" : userdetails on req ctx", value)
		var sessionDetails models.Session
		/**
			if session details not found on context
			(should be added to context by require auth middleware after
			retrieving from redis)
		**/
		if len(value) == 0 {
			// initialize session details to saved user auth
			var user models.User
			err := userCollection.FindOne(ctx, bson.M{"id": userId}).Decode(&user)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					responses.APIResponse{
						Status:    http.StatusNotFound,
						Message:   "Not found",
						Timestamp: time.Now(),
						Data:      gin.H{"error": "No user found with Id " + userId},
						Success:   false,
					},
				)
				return
			}
			sessionDetails = user.Auth
		} else {
			err := json.Unmarshal([]byte(value), &sessionDetails)
			if err != nil {
				util.ErrorLog.Println(tag+": could not parse session details", err.Error())
				c.JSON(
					http.StatusInternalServerError,
					responses.APIResponse{
						Status:    http.StatusInternalServerError,
						Message:   "Something went wrong",
						Timestamp: time.Now(),
						Data:      gin.H{"error": "Internal Error, please try again"},
						Success:   false,
					},
				)
				return
			}
		}
		util.InfoLog.Println(tag+": parsed session details ", sessionDetails)

		var operationErrorMutex sync.Mutex
		var operationError error
		var wg sync.WaitGroup
		ch := make(chan responses.TopItemsResponse, 2)
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := spotifyService.GetUserTopItems(sessionDetails.AccessToken, "tracks")
			if err != nil {
				util.ErrorLog.Println(tag+": could not retrieve user top tracks", err.Error())
				operationErrorMutex.Lock()
				operationError = err
				operationErrorMutex.Unlock()
				return
			}
			ch <- *data
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := spotifyService.GetUserTopItems(sessionDetails.AccessToken, "artists")
			if err != nil {
				util.ErrorLog.Println(tag+": could not retrieve user top artists", err.Error())
				operationErrorMutex.Lock()
				operationError = err
				operationErrorMutex.Unlock()
				return
			}
			ch <- *data
		}()

		if operationError != nil {
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": "Internal Error, please try again"},
					Success:   false,
				},
			)
			return
		}

		wg.Wait()
		close(ch)
		var userTopTracks responses.TopItemsResponse
		var userTopArtists responses.TopItemsResponse
		for data := range ch {
			if data.Items[0].Type == "artist" {
				userTopArtists = data
			} else {
				userTopTracks = data
			}
		}

		var tracksToAnalyze []string = []string{}
		for index, track := range userTopTracks.Items {
			if index < CUTOFF-1 {
				tracksToAnalyze = append(tracksToAnalyze, track.Id)
			} else {
				break
			}
		}

		features, err := spotifyService.GetTracksAudioFeatures(tracksToAnalyze, sessionDetails.AccessToken)
		if err != nil {
			util.ErrorLog.Println(tag+": could not get audio features for tracks", err.Error())
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": "Internal Error, please try again"},
					Success:   false,
				},
			)
			return
		}
		recommendationConfig := calculateRecommendationConfig(*features)
		recommendationConfig.Limit = 25
		recommendationConfig.CreatorId = userId

		var seedTracks = userTopTracks.Items[:3]
		var seedArtists = userTopArtists.Items[:2]

		var trackSeedIds []string = []string{}
		var artistSeedIds []string = []string{}
		for _, track := range seedTracks {
			trackSeedIds = append(trackSeedIds, track.Id)
		}
		for _, artist := range seedArtists {
			artistSeedIds = append(artistSeedIds, artist.Id)
		}

		recommendationConfig.SeedTracks = trackSeedIds
		recommendationConfig.SeedArtists = artistSeedIds

		recomms, err := spotifyService.GetRecommendations(sessionDetails.AccessToken, *recommendationConfig)
		if err != nil {
			util.ErrorLog.Println(tag+": could not get recommendations", err.Error())
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": "Internal Error, please try again"},
					Success:   false,
				},
			)
			return
		}

		createdPlaylist, err := spotifyService.CreatePlaylist(
			sessionDetails.AccessToken,
			userId,
			"Nubari radio for you",
			"A custom playlist built just for you",
		)
		uris := []string{}

		for _, track := range recomms.Tracks {
			uris = append(uris, track.Uri)
		}
		snapshotId, err := spotifyService.AddTracksToPlaylist(
			sessionDetails.AccessToken,
			createdPlaylist.Id,
			uris,
		)

		recommendationConfig.SnapshotId = snapshotId
		recommendationConfig.PlaylistName = "Nubari radio for you"

		if err != nil {
			util.ErrorLog.Println(tag+": could not create playlist", err.Error())
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": "Internal Error, please try again"},
					Success:   false,
				},
			)
			return
		}

		_, insertErr := recommendationProfileCollection.InsertOne(ctx, recommendationConfig)
		if insertErr != nil {
			util.ErrorLog.Println(tag+": DB Insertion err", err.Error())
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": "Internal Error, please try again"},
					Success:   false,
				},
			)
			return
		}

		c.JSON(http.StatusOK, responses.APIResponse{
			Status:    http.StatusOK,
			Message:   "Success",
			Timestamp: time.Now(),
			Data: gin.H{
				"playlist":             createdPlaylist,
				"recommendations":      recomms,
				"topTracks":            userTopTracks,
				"topArtists":           userTopArtists,
				"features":             features,
				"recommendationConfig": recommendationConfig,
				"snapshotId":           snapshotId,
			},
			Success: true,
		})
	}
}

func calculateRecommendationConfig(trackFeatures responses.TracksAudioFeatures) *models.RecommendationProfile {
	var accousticnessLow []float32
	var accousticnessHigh []float32
	var danceabilityLow []float32
	var danceabilityHigh []float32
	var energyLow []float32
	var energyHigh []float32
	var instrumentalnessLow []float32
	var instrumentalnessHigh []float32
	var livenessHigh []float32
	var livenessLow []float32
	var tempo []float32
	var valenceLow []float32
	var valenceHigh []float32

	var desiredAcousticness float32
	var desiredDanceability float32
	var desiredEnergy float32
	var desiredInstrumentalness float32
	var desiredLiveness float32
	var desiredTempo float32
	var desiredValence float32

	for _, feature := range trackFeatures.AudioFeatures {
		if feature.Acousticness >= 0.5 {
			accousticnessHigh = append(accousticnessHigh, feature.Acousticness)
		} else {
			accousticnessLow = append(accousticnessLow, feature.Acousticness)
		}

		if feature.Danceability >= 0.5 {
			danceabilityHigh = append(danceabilityHigh, feature.Danceability)
		} else {
			danceabilityLow = append(danceabilityLow, feature.Danceability)
		}

		if feature.Energy >= 0.5 {
			energyHigh = append(energyHigh, feature.Energy)
		} else {
			energyLow = append(energyLow, feature.Energy)
		}

		if feature.Instrumentalness >= 0.5 {
			instrumentalnessHigh = append(instrumentalnessHigh, feature.Instrumentalness)
		} else {
			instrumentalnessLow = append(instrumentalnessLow, feature.Instrumentalness)
		}

		if feature.Liveness >= 0.5 {
			livenessHigh = append(livenessHigh, feature.Liveness)
		} else {
			livenessLow = append(livenessLow, feature.Liveness)
		}
		tempo = append(tempo, feature.Tempo)

		if feature.Valence >= 0.5 {
			valenceHigh = append(valenceHigh, feature.Valence)
		} else {
			valenceLow = append(valenceLow, feature.Valence)
		}
	}

	if len(accousticnessLow) > len(accousticnessHigh) {
		desiredAcousticness = getAverageValue(accousticnessLow)
	} else {
		desiredAcousticness = getAverageValue(accousticnessHigh)
	}
	if len(danceabilityLow) > len(danceabilityHigh) {
		desiredDanceability = getAverageValue(danceabilityLow)
	} else {
		desiredDanceability = getAverageValue(danceabilityHigh)
	}
	if len(energyLow) > len(energyHigh) {
		desiredEnergy = getAverageValue(energyLow)
	} else {
		desiredEnergy = getAverageValue(energyHigh)
	}
	if len(instrumentalnessLow) > len(instrumentalnessHigh) {
		desiredInstrumentalness = getAverageValue(instrumentalnessLow)
	} else {
		desiredInstrumentalness = getAverageValue(instrumentalnessHigh)
	}
	if len(livenessLow) > len(livenessHigh) {
		desiredLiveness = getAverageValue(livenessLow)
	} else {
		desiredLiveness = getAverageValue(livenessHigh)
	}

	desiredTempo = getAverageValue(tempo)

	if len(valenceLow) > len(valenceHigh) {
		desiredValence = getAverageValue(valenceLow)
	} else {
		desiredValence = getAverageValue(valenceHigh)
	}

	recommendationConfig := models.RecommendationProfile{
		Acousticness:     desiredAcousticness,
		Danceability:     desiredDanceability,
		Energy:           desiredEnergy,
		Instrumentalness: desiredInstrumentalness,
		Liveness:         desiredLiveness,
		Valence:          desiredValence,
		Tempo:            desiredTempo,
		SeedArtists:      []string{},
		SeedGenres:       []string{},
		SeedTracks:       []string{},
	}

	return &recommendationConfig
}

func getAverageValue(values []float32) float32 {
	if len(values) == 0 {
		return 0.0
	}

	var sum float32 = 0.0
	for _, value := range values {
		sum += value
	}

	return sum / float32(len(values))
}
