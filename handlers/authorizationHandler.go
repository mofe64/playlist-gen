package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"mofe64/playlistGen/config"
	"mofe64/playlistGen/data/responses"
	"mofe64/playlistGen/util"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
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
		url := "https://accounts.spotify.com/api/token"
		req, err := http.NewRequest("POST", url, payload)
		if err != nil {
			util.ErrorLog.Println("Error creating request", err)
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Error creating request",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
				},
			)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", authHeader)

		resp, err := client.Do(req)
		if err != nil {
			util.ErrorLog.Println("Error executing request", err)
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Error executing request",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
				},
			)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			util.ErrorLog.Println("Error decoding response body", err)
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
				},
			)
			return
		}
		var accessTokenResponse responses.AccessTokenResponse
		err = json.Unmarshal(body, &accessTokenResponse)
		if err != nil {
			util.ErrorLog.Println("Error unmarshalling res body ", err)
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
				},
			)
			return
		}

		if resp.StatusCode != http.StatusOK {
			util.InfoLog.Println("Request returned non ok res ", err)
			c.JSON(
				resp.StatusCode,
				responses.APIResponse{
					Status:    http.StatusBadRequest,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": accessTokenResponse},
				},
			)
			return
		}
		c.JSON(resp.StatusCode, responses.APIResponse{
			Status:    resp.StatusCode,
			Message:   "Success",
			Timestamp: time.Now(),
			Data: gin.H{
				"auth": accessTokenResponse,
			},
		})

	}
}
