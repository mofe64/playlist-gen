package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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
	"github.com/sethvargo/go-retry"
	"go.mongodb.org/mongo-driver/bson"
)

var spotifyBaseAuthUrl = "https://accounts.spotify.com"
var spotifyRedirectUri = "http://localhost:5000/api/v1/auth/auth_code_callback"
var spotifyService service.SpotifyService = service.NewSpotifyService(spotifyRedirectUri)
var userCollection = config.GetCollection(config.DATABASE, "users")
var validate *validator.Validate = validator.New()
var redis = config.RedisClient

func GetAccessToken() gin.HandlerFunc {
	tag := "AUTH_HANDLER_GET_ACCESS_TOKEN_CLIENT_CRED"
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		spotifyResponse, err := spotifyService.GetAccessTokenWithClientCredentials()

		if err != nil {
			if authError, ok := err.(util.ApplicationAuthError); ok {
				util.ErrorLog.Println(tag+": Spotify get access token with client cred error ", authError)
				util.GenerateBadRequestResponse(c, authError.Message)
				return
			} else {
				util.ErrorLog.Println(tag+": Internal server error ", err.Error())
				util.GenerateInternalServerErrorResponse(c, "Something went wrong please try again")
				return
			}
		}
		util.GenerateJSONResponse(c, http.StatusOK, "", gin.H{
			"access_token": spotifyResponse.AccessToken,
			"token_type":   spotifyResponse.TokenType,
			"expires_in":   spotifyResponse.ExpiresIn,
		})
	}
}

func prepareRedirectURI() string {
	// Prepare the query parameters.
	params := url.Values{}
	params.Add("client_id", config.EnvSpotifyClientId())
	params.Add("response_type", "code")
	params.Add("redirect_uri", "http://localhost:5000/api/v1/auth/auth_code_callback")
	params.Add("state", generateRandomString(16))
	params.Add("scope", "playlist-read-private playlist-read-collaborative playlist-modify-private playlist-modify-public user-top-read user-read-recently-played user-library-modify user-library-read user-read-private user-read-email")
	baseUrl := spotifyBaseAuthUrl + "/authorize"
	redirectURL := baseUrl + "?" + params.Encode()
	return redirectURL

}

func PrepareAuthCodeURI() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		redirectURL := prepareRedirectURI()

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
	tag := "AUTHORIZATION_CODE_CALLBACK"
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		// extract and log our query parameters
		code := c.Query("code")
		state := c.Query("state")
		errorMsg := c.Query("error")
		util.InfoLog.Println(tag+": code -> ", code)
		util.InfoLog.Println(tag+": state -> ", state)
		util.ErrorLog.Println(tag+": error -> ", errorMsg)

		/**
			If our callback is triggered with an error query param,
			this means our spotify authentcation was not successful.
			We return an unauthorized response
		**/
		if len(errorMsg) != 0 {
			util.GenerateBadRequestResponse(c, errorMsg)
			return
		}

		/**
			Getting to this point means our spotify authentication was successful
			and we recieved an authorization code. We will exchange this auth code
			for an acceess and refesh token from spotify
		**/
		resp, err := spotifyService.GetAccessTokenWithAuthCode(code)

		// return status 500 and log error, if access token operation returns an error
		if err != nil {
			if authError, ok := err.(util.ApplicationAuthError); ok {
				util.ErrorLog.Println(tag+": Spotify get auth code error", err.Error())
				util.GenerateBadRequestResponse(c, authError.Message)
				return
			} else {
				util.ErrorLog.Println(tag+": Internal server error", err.Error())
				util.GenerateInternalServerErrorResponse(c, "Something went wrong, please try again")
				return
			}
		}

		/**
			Retrive the authenticated user's profile from spotify
		**/
		userProfile, err := spotifyService.GetUserProfile(resp.AccessToken)

		if err != nil {
			if authError, ok := err.(util.ApplicationAuthError); ok {
				util.ErrorLog.Println(tag+": authentication error ", authError.Message)
				// ask client to retry authorization
				redirectURL := prepareRedirectURI()
				util.GenerateJSONResponse(c, http.StatusTeapot, "Re-authorization required", gin.H{
					"url": redirectURL,
				})
				return
			}
			if rateLimitError, ok := err.(util.ApplicationRateLimitError); ok {
				util.ErrorLog.Println(tag+": Rate limit error ", rateLimitError.Message)
				// retry and back off strat.
				backOff := retry.NewFibonacci(3 * time.Second)
				backOff = retry.WithMaxRetries(3, backOff)

				if err := retry.Do(ctx, backOff, func(_ context.Context) error {
					userProfile, err = spotifyService.GetUserProfile(resp.AccessToken)
					if err != nil {
						if rateLimitError, ok := err.(util.ApplicationRateLimitError); ok {
							return retry.RetryableError(rateLimitError)
						} else {
							util.GenerateInternalServerErrorResponse(c, "Something went wrong please try again")
							return nil
						}
					}
					return nil
				}); err != nil {
					util.GenerateInternalServerErrorResponse(c, "Something went wrong please try again")
				}
				return
			}
			if applicationError, ok := err.(util.ApplicationError); ok {
				util.ErrorLog.Println(tag+": Internal server error ", applicationError.Message)
				util.GenerateInternalServerErrorResponse(c, "Something went wrong please try again")
				return
			}
		}

		util.InfoLog.Println(tag+": user profile ", userProfile)

		var user models.User
		/**
			Create a new Session struct called spotify auth which will hold our
			spotify authentication details
		**/
		var spotifyAuth = models.Session{
			AccessToken:  resp.AccessToken,
			TokenType:    resp.TokenType,
			Scope:        resp.Scope,
			RefreshToken: resp.RefreshToken,
			ExpiresIn:    resp.ExpiresIn,
			IssuedAt:     time.Now(),
		}

		/**
			Check if the retrieved profile exists in our database
			If the user does not exist, create a new entry in database for user
			If the user exists, update user's authentication field to reflect newly
			obtained access and refresh tokens
		**/
		retrieveError := userCollection.FindOne(ctx, bson.M{"id": userProfile.Id}).Decode(&user)
		// if user does not exist
		if retrieveError != nil {
			util.InfoLog.Println(tag+": retrieve error", retrieveError.Error())
			// create new user entry in db based off retrived spotify profile and
			// authentication details
			newUser := models.User{
				Id:          userProfile.Id,
				Username:    userProfile.DisplayName,
				Country:     userProfile.Country,
				Email:       userProfile.Email,
				SpotifyPlan: userProfile.Product,
				Auth:        spotifyAuth,
			}
			_, err := userCollection.InsertOne(ctx, newUser)
			if err != nil {
				util.ErrorLog.Println(tag+": DB Insertion err", err.Error())
				util.GenerateInternalServerErrorResponse(c, "Something went wrong, please try again")
				return
			}
			user = newUser
		} else {
			// if user exists
			// update user's auth details to the newly generated auth details
			user.Auth = spotifyAuth
			filter := bson.M{"id": user.Id}
			update := bson.M{"$set": user}
			_, err := userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				util.ErrorLog.Println(tag+": DB Update err", err.Error())
				util.GenerateInternalServerErrorResponse(c, "Something went wrong, please try again")
				return
			}

		}
		/**
			To avoid having to go to the db everytime to retrieve user's auth details,
			we are going to store the auth details in redis, mapped to the user's id
			This way we can more efficiently retrieve the auth details and use it.
			To do this, we first convert our session struct (spotifyAuth) to json,
			we will then stringify the json value and store it as a string in redis.
			The auth details will be mapped to the user id for easy retrieval
		**/
		jsonValue, err := json.Marshal(spotifyAuth)
		// if there was an error converting auth details to json format return 500 error res
		if err != nil {
			util.ErrorLog.Println(tag+": Could not convert spotify token details to json", err.Error())
			util.GenerateInternalServerErrorResponse(c, "Something went wrong, please try again")
			return
		}
		// store spotify auth details in redis kv
		// Todo handle siuations where redis is unavailable
		util.InfoLog.Println(tag + ": about to set to redis key --> " + user.Id)
		redis.Set(ctx, user.Id, string(jsonValue), 0)

		// return success response to user
		util.GenerateJSONResponse(c, http.StatusOK, "Success", gin.H{
			"access_token": resp.AccessToken,
			"token_type":   resp.TokenType,
			"expires_in":   resp.ExpiresIn,
			"userId":       user.Id,
		})
	}
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}
