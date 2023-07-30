package middleware

import (
	"mofe64/playlistGen/data/responses"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		const bearerSchema = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.APIResponse{
				Status:    http.StatusUnauthorized,
				Message:   "Authorization Header missing ...",
				Timestamp: time.Now(),
				Data:      gin.H{},
				Success:   false,
			})
		}

		c.Next()
	}

}
