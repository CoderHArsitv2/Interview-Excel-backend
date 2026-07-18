package routes

import (
	"interviewexcel-backend-go/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/auth/register", controllers.Signup)
	router.POST("/auth/signin", controllers.UserSignIn)
	router.POST("/auth/google/login", controllers.UserGoogleAuth)

	router.POST("/auth/user", controllers.GetUser)
	router.GET("/auth/refresh", controllers.RefreshSession)
}
