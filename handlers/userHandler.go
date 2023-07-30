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
)

func CreatePlaylist() gin.HandlerFunc {
	tag := "CREATE_PLAYLIST_HANDLER"
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		userId := c.Param("userId")
		value := c.GetString("userDetails")
		var tokenDetails models.Token
		if len(value) == 0 {
			// retrieve token details from db
		} else {
			err := json.Unmarshal([]byte(value), &tokenDetails)
			if err != nil {
				util.ErrorLog.Println(tag+": could not parse auth token", value)
			}
		}

		util.InfoLog.Println(tag+": user Id", userId)
		c.JSON(http.StatusOK, responses.APIResponse{
			Status:    http.StatusOK,
			Message:   "Success",
			Timestamp: time.Now(),
			Data:      gin.H{},
			Success:   true,
		})
	}
}
