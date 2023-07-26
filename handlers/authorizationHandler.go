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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var spotifyBaseAuthUrl = "https://accounts.spotify.com"
var spotifyRedirectUri = "http://localhost:5000/api/v1/auth/auth_code_callback"
var spotifyService service.SpotifyService = service.NewSpotifyService(spotifyRedirectUri)
var userCollection = config.GetCollection(config.DATABASE, "users")
var validate *validator.Validate = validator.New()

func GetAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		spotifyResponse, err := spotifyService.GetAccessTokenWithClientCredentials()

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
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
		params.Add("scope", "playlist-read-private playlist-read-collaborative playlist-modify-private playlist-modify-public user-top-read user-read-recently-played user-library-modify user-library-read user-read-private")
		baseUrl := spotifyBaseAuthUrl + "/authorize"
		redirectURL := baseUrl + "?" + params.Encode()
		c.Redirect(http.StatusFound, redirectURL)
	}
}

func AuthorizationCodeCallBack() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		code := c.Query("code")
		state := c.Query("state")
		error := c.Query("error")
		util.InfoLog.Println("code -> ", code)
		util.InfoLog.Println("state -> ", state)
		util.InfoLog.Println("error -> ", error)

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
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Something went wrong",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
					Success:   false,
				},
			)
			return
		}
		messageValue := "Fail"
		if resp.StatusCode == 200 {
			messageValue = "Success"
		}

		c.JSON(http.StatusOK, responses.APIResponse{
			Status:    http.StatusOK,
			Message:   messageValue,
			Timestamp: time.Now(),
			Data: gin.H{
				"auth": resp,
			},
			Success: resp.StatusCode == 200,
		})
	}
}

func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			util.ErrorLog.Printf("Error binding Json Obj %v\n", err)
			c.JSON(
				http.StatusBadRequest,
				responses.APIResponse{
					Status:    http.StatusBadRequest,
					Message:   "Invalid request",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
					Success:   false,
				},
			)
			return
		}

		if validationErr := validate.Struct(&user); validationErr != nil {
			util.ErrorLog.Println("validation error", validationErr.Error())
			c.JSON(http.StatusBadRequest,
				responses.APIResponse{
					Status:    http.StatusBadRequest,
					Message:   "validation error",
					Timestamp: time.Now(),
					Data:      gin.H{"data": validationErr.Error()},
					Success:   false,
				})
			return
		}

		var password = user.Password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			util.ErrorLog.Println("password hash error", err.Error())
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "Failed to hash password",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
					Success:   false,
				},
			)
			return
		}

		newUser := models.User{
			Id:       primitive.NewObjectID(),
			Username: user.Username,
			Email:    user.Email,
			Password: string(hashedPassword),
		}

		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			util.ErrorLog.Println("DB Insertion err", err.Error())
			c.JSON(
				http.StatusInternalServerError,
				responses.APIResponse{
					Status:    http.StatusInternalServerError,
					Message:   "DB Insertion err",
					Timestamp: time.Now(),
					Data:      gin.H{"error": err.Error()},
					Success:   false,
				},
			)
			return
		}
		c.JSON(http.StatusCreated, responses.APIResponse{
			Status:  http.StatusCreated,
			Message: "success",
			Data:    gin.H{"data": result},
		})

	}

}

func LoginUser() {}

func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}
