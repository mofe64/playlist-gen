package routes

import (
	"mofe64/playlistGen/handlers"

	"github.com/gin-gonic/gin"
)

func AuthorizationRoute(router *gin.Engine) {
	authorizationRoutes := router.Group("api/v1/auth")
	{
		authorizationRoutes.GET("/client_cred", handlers.GetAccessToken())
	}
}
