package middleware

import (
	"context"
	"mofe64/playlistGen/config"
	"mofe64/playlistGen/data/responses"
	"mofe64/playlistGen/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var redis = config.RedisClient
var tag = "REQUIRE_AUTH_MIDDLEWARE"

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		userId := c.Param("userId")
		if len(userId) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.APIResponse{
				Status:    http.StatusUnauthorized,
				Message:   "User Id path variable required",
				Timestamp: time.Now(),
				Data:      gin.H{},
				Success:   false,
			})
			return
		}
		// update redis config creation to ping redis and panic not here
		if err := redis.Ping(ctx).Err(); err != nil {
			util.ErrorLog.Println(tag+": could not reach redis ...", err.Error())
		} else {
			val, err := redis.Get(ctx, userId).Result()
			if err != nil {
				util.ErrorLog.Println(tag+": Could not retrive from redis", err.Error())
			} else {
				c.Set("userDetails", val)
			}
		}

		c.Next()
	}

}
