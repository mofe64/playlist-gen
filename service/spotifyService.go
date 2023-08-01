package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mofe64/playlistGen/config"
	"mofe64/playlistGen/data/responses"
	"mofe64/playlistGen/util"
	"net/http"
	"net/url"
	"strings"
)

type SpotifyService interface {
	GetAccessTokenWithClientCredentials() (*responses.AccessTokenResponse, error)
	GetAccessTokenWithAuthCode(authCode string) (*responses.AccessTokenResponse, error)
	GetUserProfile(accessToken string) (*responses.SpotifyUserProfile, error)
	GetUserTopTracks(accessToken string) (*responses.TopItemsResponse, error)
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

	var accessTokenResponse responses.AccessTokenResponse
	err = json.Unmarshal(body, &accessTokenResponse)
	if err != nil {
		util.ErrorLog.Println("Error unmarshalling res body ", err)
		return nil, err
	}
	accessTokenResponse.StatusCode = resp.StatusCode
	return &accessTokenResponse, nil

}

func (s *spotifyService) GetUserProfile(accesstoken string) (*responses.SpotifyUserProfile, error) {
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
	var userProfile responses.SpotifyUserProfile
	err = json.Unmarshal(body, &userProfile)
	if err != nil {
		util.ErrorLog.Println("Error unmarshalling res body ", err)
		return nil, err
	}

	return &userProfile, nil

}

func (s *spotifyService) GetUserTopTracks(accessToken string) (*responses.TopItemsResponse, error) {
	var tag = "SPOTIFY_SERVICE_GET_USER_TOP_TRACKS"
	requestBaseUrl := s.spotifyBaseWebApi + "/me/top/tracks"
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

func (s *spotifyService) GetAccessTokenWithAuthCode(authCode string) (*responses.AccessTokenResponse, error) {

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
	var accessTokenResponse responses.AccessTokenResponse
	err = json.Unmarshal(body, &accessTokenResponse)
	if err != nil {
		util.ErrorLog.Println("Error unmarshalling res body ", err)
		return nil, err
	}
	accessTokenResponse.StatusCode = resp.StatusCode
	return &accessTokenResponse, nil
}
