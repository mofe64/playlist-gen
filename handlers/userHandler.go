package handlers

import (
	"context"
	"encoding/json"
	"mofe64/playlistGen/data/models"
	"mofe64/playlistGen/data/responses"
	"mofe64/playlistGen/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func CreatePlaylist() gin.HandlerFunc {
	tag := "CREATE_PLAYLIST_HANDLER"
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		userId := c.Param("userId")
		value := c.GetString("userDetails")
		util.InfoLog.Println(tag+" : userdetails on req ctx", value)
		var sessionDetails models.Session
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
		userTopTracks, err := spotifyService.GetUserTopTracks(sessionDetails.AccessToken)
		if err != nil {
			util.ErrorLog.Println(tag+": could not retrieve user top tracks", err.Error())
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
				"topTracks": userTopTracks,
			},
			Success: true,
		})
	}
}
