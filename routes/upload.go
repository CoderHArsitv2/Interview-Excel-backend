package routes

import (
	"interviewexcel-backend-go/controllers"
	"interviewexcel-backend-go/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUploadRoutes(router *gin.Engine) {
	uploadGroup := router.Group("/upload")
	uploadGroup.Use(middleware.AuthMiddleware())

	uploadGroup.POST("/image", controllers.UploadImage)
}
