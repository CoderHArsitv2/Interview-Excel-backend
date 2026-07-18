package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	logger "interviewexcel-backend-go/pkg/errors"
	"interviewexcel-backend-go/utils"
	"net/http"
	"os"
	"strings"
	"time"
)

type Claims struct {
	UserID string `json:"user_uuid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			logger.Error("token is empty")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Token"})
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		//Check if the token is BlacListed
		isBlacklisted, err := utils.IsTokenBlacklisted(tokenString)
		if err != nil {
			logger.Error("error in getting token isblacklisting:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server Error"})
			return
		}
		if isBlacklisted {
			logger.Error("unauthorizedn")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		}
		// Parse and validate the JWT token
		var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			logger.Error("error in parsing the token:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			// Ensure the token hasn't expired (already handled by jwt.RegisteredClaims)
			if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
				logger.Error("Token has expired")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
				return
			}

			// Pass the user ID or other relevant claims to the next handlers
			c.Set("user_uuid", claims.UserID)
			c.Set("role", claims.Role)
			c.Next()
		} else {
			logger.Error("Invalid Token Claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}
