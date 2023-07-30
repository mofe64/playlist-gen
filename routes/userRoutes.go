package routes

import (
	"mofe64/playlistGen/handlers"
	"mofe64/playlistGen/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine) {
	userRoutes := router.Group("api/v1/user", middleware.RequireAuth())
	{
		userRoutes.GET("/:userId/create_playlist", handlers.CreatePlaylist())
	}
}
