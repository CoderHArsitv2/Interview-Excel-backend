package controllers

import (
	"interviewexcel-backend-go/config"
	"interviewexcel-backend-go/models"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	logger "interviewexcel-backend-go/pkg/errors"
)

type AccessClaims struct {
	UserID string `json:"user_uuid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GetUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		logger.Error("auth header is empty")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing access token"})
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and verify token
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		logger.Error("error in parsing token:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
		return
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token claims"})
		return
	}

	// Fetch user
	userRepo := models.InitUserRepo(config.DB)
	user, err := userRepo.GetByUUID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return only ID and role
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":   user.ID,
			"uuid": user.UserUUID,
			"role": user.Role,
		},
	})
}
