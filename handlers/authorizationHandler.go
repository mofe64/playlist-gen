package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"mofe64/playlistGen/config"
	"mofe64/playlistGen/data/models"
	"mofe64/playlistGen/data/responses"
	"mofe64/playlistGen/service"
	"mofe64/playlistGen/util"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
)

var spotifyBaseAuthUrl = "https://accounts.spotify.com"
var spotifyRedirectUri = "http://localhost:5000/api/v1/auth/auth_code_callback"
var spotifyService service.SpotifyService = service.NewSpotifyService(spotifyRedirectUri)
var userCollection = config.GetCollection(config.DATABASE, "users")
var validate *validator.Validate = validator.New()
var redis = config.RedisClient

func GetAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		spotifyResponse, err := spotifyService.GetAccessTokenWithClientCredentials()

		if err != nil {
			util.ErrorLog.Println("Spotify get access token with client cred error ", err.Error())
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

		go func() {
			applicationContext := util.GetApplicationContextInstance()
			applicationContext.SetAccessToken(spotifyResponse.AccessToken)
		}()

		messageValue := "Fail"
		if spotifyResponse.StatusCode == 200 {
			messageValue = "Success"
		}

		c.JSON(spotifyResponse.StatusCode, responses.APIResponse{
			Status:    spotifyResponse.StatusCode,
			Message:   messageValue,
			Timestamp: time.Now(),
			Data: gin.H{
				"auth": spotifyResponse,
			},
			Success: spotifyResponse.StatusCode == 200,
		})

	}
}

func GetAuthorizationCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Prepare the query parameters.
		params := url.Values{}
		params.Add("client_id", config.EnvSpotifyClientId())
		params.Add("response_type", "code")
		params.Add("redirect_uri", "http://localhost:5000/api/v1/auth/auth_code_callback")
		params.Add("state", generateRandomString(16))
		params.Add("scope", "playlist-read-private playlist-read-collaborative playlist-modify-private playlist-modify-public user-top-read user-read-recently-played user-library-modify user-library-read user-read-private user-read-email")
		baseUrl := spotifyBaseAuthUrl + "/authorize"
		redirectURL := baseUrl + "?" + params.Encode()
		// c.Redirect(http.StatusFound, redirectURL)
		c.JSON(http.StatusOK, responses.APIResponse{
			Status:    http.StatusOK,
			Message:   "Success",
			Timestamp: time.Now(),
			Data: gin.H{
				"url": redirectURL,
			},
			Success: true,
		})
	}
}

func AuthorizationCodeCallBack() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		code := c.Query("code")
		state := c.Query("state")
		error := c.Query("error")
		util.InfoLog.Println("code -> ", code)
		util.InfoLog.Println("state -> ", state)
		util.ErrorLog.Println("error -> ", error)

		if len(error) != 0 {
			c.JSON(http.StatusUnauthorized, responses.APIResponse{
				Status:    http.StatusUnauthorized,
				Message:   error,
				Timestamp: time.Now(),
				Data:      gin.H{},
				Success:   false,
			})
			return
		}
		resp, err := spotifyService.GetAccessTokenWithAuthCode(code)

		if err != nil {
			util.ErrorLog.Println("Spotify get auth code error", err.Error())
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

		userProfile, err := spotifyService.GetUserProfile(resp.AccessToken)
		util.InfoLog.Println("user profile ", userProfile)
		// check if user exists in db if he does not save
		var user models.User
		retrieveError := userCollection.FindOne(ctx, bson.M{"id": userProfile.Id}).Decode(&user)
		if retrieveError != nil {
			util.InfoLog.Println("retrieve error", retrieveError.Error())
			newUser := models.User{
				Id:          userProfile.Id,
				Username:    userProfile.DisplayName,
				Country:     userProfile.Country,
				Email:       userProfile.Email,
				SpotifyPlan: userProfile.Product,
			}
			_, err := userCollection.InsertOne(ctx, newUser)
			if err != nil {
				util.ErrorLog.Println("DB Insertion err", err.Error())
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
		messageValue := "Fail"
		if resp.StatusCode == 200 {
			messageValue = "Success"
		}

		var userSpotifyTokenDetails = models.Token{
			AccessToken:  resp.AccessToken,
			TokenType:    resp.TokenType,
			Scope:        resp.Scope,
			RefreshToken: resp.RefreshToken,
		}

		redis.Set(ctx, user.Id, userSpotifyTokenDetails, 0)

		c.JSON(http.StatusOK, responses.APIResponse{
			Status:    http.StatusOK,
			Message:   messageValue,
			Timestamp: time.Now(),
			Data: gin.H{
				"access_token": resp.AccessToken,
				"token_type":   resp.TokenType,
				"expires_in":   resp.ExpiresIn,
				"userId":       user.Id,
			},
			Success: resp.StatusCode == 200,
		})
	}
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}
