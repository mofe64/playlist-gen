package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mofe64/playlistGen/config"
	"mofe64/playlistGen/data/models"
	"mofe64/playlistGen/data/responses"
	"mofe64/playlistGen/util"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type SpotifyService interface {
	GetAccessTokenWithClientCredentials() (*responses.AccessTokenResponse, error)
	GetAccessTokenWithAuthCode(authCode string) (*responses.AccessTokenResponse, error)
	GetUserProfile(accessToken string) (*responses.SpotifyUserProfile, error)
	GetUserTopItems(accessToken string, entityType string) (*responses.TopItemsResponse, error)
	GetTracksAudioFeatures(trackIds []string, accessToken string) (*responses.TracksAudioFeatures, error)
	GetRecommendations(accessToken string, config models.RecommendationProfile) (*responses.RecommendationsResponse, error)
	CreatePlaylist(accessToken string, userId string, name string, desc string) (*models.Playlist, error)
	AddTracksToPlaylist(accessToken string, playlistId string, trackUris []string) (string, error)
}

type spotifyService struct {
	spotifyBaseAuthUrl string
	spotifyRedirectUri string
	spotifyBaseWebApi  string
}

func NewSpotifyService(redirectUri string) SpotifyService {
	return &spotifyService{
		spotifyBaseAuthUrl: "https://accounts.spotify.com",
		spotifyRedirectUri: redirectUri,
		spotifyBaseWebApi:  "https://api.spotify.com/v1",
	}
}

func (s *spotifyService) GetAccessTokenWithClientCredentials() (*responses.AccessTokenResponse, error) {
	tag := "SPOTIFY_SERVICE_GET_ACCESS_TOKEN_CLIENT_CRED"
	/**
		Since we are sending formData to the spotify api as application/x-www-form-urlencoded format, we use
		formData := url.Values{} to create an empty url.Values struct.
		url.Values is a type provided by the net/url package in Go,It is typically used for query parameters and form values
		and it is used to represent URL query parameters in a key-value format.
	**/
	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")

	/**
		formData.Encode() encodes the URL query parameters in the formData url.Values struct into
		the application/x-www-form-urlencoded format. This encoding is commonly used for encoding form data in HTTP POST requests.
		The data.Encode() method returns a URL-encoded string representation of the key-value pairs in the url.Values struct.
		It takes care of properly escaping special characters and converting spaces to the + sign
		as required in the application/x-www-form-urlencoded format.
	**/
	payload := strings.NewReader(formData.Encode())

	clientId := config.EnvSpotifyClientId()
	clientSecret := config.EnvSpotifyClientSecret()
	authString := clientId + ":" + clientSecret

	/**
		[]byte(authString) converts the authString (which is a string containing the client ID and client secret concatenated with a colon)
		into a byte slice, since Base64 encoding works with byte slices.
		base64.StdEncoding.EncodeToString() takes the byte slice as input and returns a Base64-encoded string representation of the input data.
		The resulting encodedAuthString will be a Base64-encoded representation of the client ID and client secret
		in the format required for the Authorization header in an HTTP request.
	**/
	encodedAuthString := base64.StdEncoding.EncodeToString([]byte(authString))

	authHeader := "Basic " + encodedAuthString
	client := &http.Client{}
	url := s.spotifyBaseAuthUrl + "/api/token"
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		util.ErrorLog.Println(tag+": Error creating request", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLog.Println(tag+": Error executing request", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.ErrorLog.Println(tag+": Error decoding response body", err)
		return nil, err
	}

	if resp.StatusCode == 400 {
		var spotifyErrorRes responses.SpotifyAuthErrorReponse
		err = json.Unmarshal(body, &spotifyErrorRes)
		if err != nil {
			util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
			return nil, err
		}
		var authError = util.ApplicationAuthError{
			Message: spotifyErrorRes.ErrorDescription,
		}
		return nil, authError
	}

	var accessTokenResponse responses.AccessTokenResponse
	err = json.Unmarshal(body, &accessTokenResponse)
	if err != nil {
		util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
		return nil, err
	}
	return &accessTokenResponse, nil

}

func (s *spotifyService) GetUserProfile(accesstoken string) (*responses.SpotifyUserProfile, error) {
	tag := "SPOTIFY_SERVICE_GET_USER_PROFILE"

	url := s.spotifyBaseWebApi + "/me"
	authHeader := "Bearer " + accesstoken
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		util.ErrorLog.Println("Error creating request", err)
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLog.Println("Error executing request", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.ErrorLog.Println("Error decoding response body", err)
		return nil, err
	}

	if resp.StatusCode == 401 {
		var operationErr responses.SpotifyOperationErrorResponse
		err = json.Unmarshal(body, &operationErr)
		if err != nil {
			util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
			return nil, err
		}
		var authError = util.ApplicationAuthError{
			Message: operationErr.Error.Message,
		}
		return nil, authError
	}

	if resp.StatusCode == 403 {
		var operationErr responses.SpotifyOperationErrorResponse
		err = json.Unmarshal(body, &operationErr)
		if err != nil {
			util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
			return nil, err
		}
		var applicationError = util.ApplicationError{
			Message: operationErr.Error.Message,
		}
		return nil, applicationError

	}

	if resp.StatusCode == 429 {
		var operationErr responses.SpotifyOperationErrorResponse
		err = json.Unmarshal(body, &operationErr)
		if err != nil {
			util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
			return nil, err
		}
		var rateLimitError = util.ApplicationRateLimitError{
			Message: operationErr.Error.Message,
		}
		return nil, rateLimitError
	}

	var userProfile responses.SpotifyUserProfile
	err = json.Unmarshal(body, &userProfile)
	if err != nil {
		util.ErrorLog.Println("Error unmarshalling res body ", err)
		return nil, err
	}

	return &userProfile, nil

}

func (s *spotifyService) AddTracksToPlaylist(accessToken string, playlistId string, trackUris []string) (string, error) {
	var tag = "SPOTIFY_SERVICE_ADD_TRACKS_TO_pLAYLIST"
	reqUrl := s.spotifyBaseWebApi + "/playlists/" + playlistId + "/tracks"
	authHeader := "Bearer " + accessToken
	util.InfoLog.Println(tag+": uris are --> ", trackUris)
	payload := map[string]interface{}{
		"uris":     trackUris,
		"position": 0,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		util.ErrorLog.Println(tag+": Error marshalling req body  ", err)
		return "", err
	}
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		util.ErrorLog.Println(tag+": Error creating req  ", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLog.Println(tag+": Error executing request", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.ErrorLog.Println(tag+": Error decoding response body", err)
		return "", err
	}
	snapshotId := string(body)
	return snapshotId, nil
}

func (s *spotifyService) CreatePlaylist(accessToken string, userId string, name string, desc string) (*models.Playlist, error) {
	var tag = "SPOTIFY_SERVICE_CREATE_PLAYLIST"
	reqUrl := s.spotifyBaseWebApi + "/users/" + userId + "/playlists"
	authHeader := "Bearer " + accessToken
	payload := map[string]interface{}{
		"name":        name,
		"description": desc,
		"public":      false,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		util.ErrorLog.Println(tag+": Error marshalling req body  ", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		util.ErrorLog.Println(tag+": Error creating req  ", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLog.Println(tag+": Error executing request", err)
		return nil, err
	}
	defer resp.Body.Close()

	// error handling strat, check status code at this point, if not success
	// return nill plus error object with descriptive message
	//  // Check the response status code
	//  if resp.StatusCode == http.StatusOK {
	// 	fmt.Println("POST request was successful!")
	// } else {
	// 	fmt.Println("POST request failed with status code:", resp.StatusCode)
	// }

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.ErrorLog.Println(tag+": Error decoding response body", err)
		return nil, err
	}

	var playlist models.Playlist
	err = json.Unmarshal(body, &playlist)
	if err != nil {
		util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
		return nil, err
	}

	return &playlist, nil
}

func (s *spotifyService) GetUserTopItems(accessToken string, entityType string) (*responses.TopItemsResponse, error) {
	var tag = "SPOTIFY_SERVICE_GET_USER_TOP_ITEMS"
	requestBaseUrl := s.spotifyBaseWebApi + "/me/top/" + entityType
	queryParams := url.Values{}
	queryParams.Set("time_range", "short_term")
	queryParams.Set("limit", "50")
	fullUrl := fmt.Sprintf("%s?%s", requestBaseUrl, queryParams.Encode())

	authHeader := "Bearer " + accessToken
	client := &http.Client{}
	req, err := http.NewRequest("GET", fullUrl, nil)

	if err != nil {
		util.ErrorLog.Println(tag+": Error creating request", err)
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLog.Println(tag+": Error executing request", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.ErrorLog.Println(tag+": Error decoding response body", err)
		return nil, err
	}
	var topItems responses.TopItemsResponse
	err = json.Unmarshal(body, &topItems)
	if err != nil {
		util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
		return nil, err
	}

	return &topItems, nil
}

func (s *spotifyService) GetRecommendations(accessToken string, config models.RecommendationProfile) (*responses.RecommendationsResponse, error) {
	var tag = "SPOTIFY_SERVICE_GET_RECOMMENDATIONS"
	requestBaseUrl := s.spotifyBaseWebApi + "/recommendations"
	queryParams := url.Values{}
	queryParams.Set("seed_artists", strings.Join(config.SeedArtists, ","))
	queryParams.Set("seed_tracks", strings.Join(config.SeedTracks, ","))
	queryParams.Set("limit", strconv.FormatInt(int64(config.Limit), 10))
	queryParams.Set("target_acousticness", strconv.FormatFloat(float64(config.Acousticness), 'f', -1, 32))
	queryParams.Set("target_danceability", strconv.FormatFloat(float64(config.Danceability), 'f', -1, 32))
	queryParams.Set("target_energy", strconv.FormatFloat(float64(config.Energy), 'f', -1, 32))
	queryParams.Set("target_instrumentalness", strconv.FormatFloat(float64(config.Instrumentalness), 'f', -1, 32))
	queryParams.Set("target_liveness", strconv.FormatFloat(float64(config.Liveness), 'f', -1, 32))
	queryParams.Set("target_tempo", strconv.FormatFloat(float64(config.Tempo), 'f', -1, 32))
	queryParams.Set("target_valence", strconv.FormatFloat(float64(config.Valence), 'f', -1, 32))
	fullUrl := fmt.Sprintf("%s?%s", requestBaseUrl, queryParams.Encode())

	authHeader := "Bearer " + accessToken
	client := &http.Client{}
	req, err := http.NewRequest("GET", fullUrl, nil)

	if err != nil {
		util.ErrorLog.Println(tag+": Error creating request", err)
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLog.Println(tag+": Error executing request", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.ErrorLog.Println(tag+": Error decoding response body", err)
		return nil, err
	}
	var recommendations responses.RecommendationsResponse
	err = json.Unmarshal(body, &recommendations)
	if err != nil {
		util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
		return nil, err
	}

	return &recommendations, nil
}

func (s *spotifyService) GetTracksAudioFeatures(trackIds []string, accessToken string) (*responses.TracksAudioFeatures, error) {
	var tag = "SPOTIFY_SERVICE_GET_TRACK_AUDIO_FEATURES"
	if len(trackIds) > 100 {
		util.ErrorLog.Println(tag + ": track ids cannot be more than 100")
		return nil, errors.New("track ids cannot be more than 100 in length")
	}
	requestUrl := s.spotifyBaseWebApi + "/audio-features"
	commaSeperatedIdString := strings.Join(trackIds, ",")
	queryParams := url.Values{}
	queryParams.Set("ids", commaSeperatedIdString)
	fullUrl := fmt.Sprintf("%s?%s", requestUrl, queryParams.Encode())

	authHeader := "Bearer " + accessToken
	client := &http.Client{}
	req, err := http.NewRequest("GET", fullUrl, nil)

	if err != nil {
		util.ErrorLog.Println(tag+": Error creating request", err)
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLog.Println(tag+": Error executing request", err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.ErrorLog.Println(tag+": Error decoding response body", err)
		return nil, err
	}

	var features responses.TracksAudioFeatures
	err = json.Unmarshal(body, &features)
	if err != nil {
		util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
		return nil, err
	}

	return &features, nil
}

func (s *spotifyService) GetAccessTokenWithAuthCode(authCode string) (*responses.AccessTokenResponse, error) {
	tag := "SPOTIFY_SERVICE_GET_ACCESS_TOKEN_WITH_AUTH_CODE"

	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", authCode)
	formData.Set("redirect_uri", s.spotifyRedirectUri)

	payload := strings.NewReader(formData.Encode())

	clientId := config.EnvSpotifyClientId()
	clientSecret := config.EnvSpotifyClientSecret()
	authString := clientId + ":" + clientSecret

	encodedAuthString := base64.StdEncoding.EncodeToString([]byte(authString))

	authHeader := "Basic " + encodedAuthString
	client := &http.Client{}
	url := s.spotifyBaseAuthUrl + "/api/token"
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		util.ErrorLog.Println("Error creating request", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLog.Println("Error executing request", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.ErrorLog.Println("Error decoding response body", err)
		return nil, err
	}

	if resp.StatusCode == 400 {
		var spotifyErrorRes responses.SpotifyAuthErrorReponse
		err = json.Unmarshal(body, &spotifyErrorRes)
		if err != nil {
			util.ErrorLog.Println(tag+": Error unmarshalling res body ", err)
			return nil, err
		}
		var authError = util.ApplicationAuthError{
			Message: spotifyErrorRes.ErrorDescription,
		}
		return nil, authError
	}

	var accessTokenResponse responses.AccessTokenResponse
	err = json.Unmarshal(body, &accessTokenResponse)
	if err != nil {
		util.ErrorLog.Println("Error unmarshalling res body ", err)
		return nil, err
	}
	return &accessTokenResponse, nil
}
