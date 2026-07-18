package controllers

import (
	"interviewexcel-backend-go/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
		return
	}
	// Add the token to the blacklist with the remaining time until it expires
	expiration := time.Hour * 24 //Replace with actual token expiration duration

	if err := utils.AddTokenToBlacklist(token, expiration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Black list token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SuccessFully logged out"})
}
