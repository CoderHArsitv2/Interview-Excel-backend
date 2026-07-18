package controllers

import (
	"interviewexcel-backend-go/config"
	"interviewexcel-backend-go/models"
	logger "interviewexcel-backend-go/pkg/errors"
	utils "interviewexcel-backend-go/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

const refreshTokenMaxAge = 7 * 24 * 60 * 60

func setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	runtimeConfig := config.RuntimeConfig()

	c.SetCookie(
		"refresh_token",
		refreshToken,
		refreshTokenMaxAge,
		"/",
		runtimeConfig.CookieDomain,
		runtimeConfig.CookieSecure,
		true,
	)
}

func persistRefreshToken(refreshToken string, value interface{}) error {
	if config.RedisClient == nil {
		return nil
	}

	return config.RedisClient.Set(
		config.Ctx,
		refreshToken,
		value,
		7*24*time.Hour,
	).Err()
}

func Signup(c *gin.Context) {
	var req SignUpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// check if user already exists
	var existing models.User
	if err := config.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		logger.Errorf("User already exists: %v\n", existing.Email)
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// hash password if provided
	var password *string
	if req.Password != "" {
		hash, _ := utils.HashPassword(req.Password)
		password = &hash
	}

	user := models.User{
		FullName: req.FullName,
		UserUUID: utils.GenerateUserUUID(req.Role),
		Email:    req.Email,
		Password: password,
		Role:     req.Role, // "student" or "expert"
	}

	if err := config.DB.Create(&user).Error; err != nil {
		logger.Errorf("Error creating user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 🔹 Create respective profile
	switch req.Role {
	case "student":
		student := models.Student{UserID: user.UserUUID}
		if err := config.DB.Create(&student).Error; err != nil {
			logger.Errorf("Error creating student profile: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student profile"})
			return
		}
	case "expert":
		expert := models.Expert{UserID: user.UserUUID}
		if err := config.DB.Create(&expert).Error; err != nil {
			logger.Errorf("Error creating expert profile: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expert profile"})
			return
		}
	}

	// 🔹 Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.UserUUID, user.Role)
	if err != nil {
		logger.Errorf("Error generating access token: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.UserUUID, user.Role)
	if err != nil {
		logger.Errorf("Error generating refresh token: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	setRefreshTokenCookie(c, refreshToken)

	// Save refresh token in Redis
	if err := persistRefreshToken(refreshToken, user.ID); err != nil {
		logger.Errorf("Error saving refresh token to Redis: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"access_token": accessToken,
	})
}

func UserGoogleAuth(c *gin.Context) {
	var req GoogleAuthRequest

	err := c.ShouldBindJSON(&req)
	if err != nil || (req.Role != "student" && req.Role != "expert") {
		logger.Error("Error in binding request, ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Google token or invalid role"})
		return
	}

	// Verify the Google token
	payload, err := idtoken.Validate(c, req.Token, "")
	if err != nil {
		logger.Error("error in verifying google token, ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google token"})
		return
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		logger.Error("error: Email not found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
		return
	}

	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)

	if !emailVerified {
		logger.Error("Email not verified by Google")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not verified by Google"})
		return
	}

	// Initialize repos
	userRepo := models.InitUserRepo(config.DB)
	studentRepo := models.InitStudentRepo(config.DB)
	expertRepo := models.InitExpertRepo(config.DB)

	// Check if user exists
	user, err := userRepo.GetByEmail(email)
	if err != nil {
		// New user
		user = &models.User{
			FullName: name,
			UserUUID: utils.GenerateUserUUID(req.Role),
			Email:    email,
			Picture:  picture,
			Role:     req.Role,
		}

		err := userRepo.Create(user)
		if err != nil {
			logger.Error("error in creating the user: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Create respective profile
		switch req.Role {
		case "student":
			err := studentRepo.Create(&models.Student{UserID: user.UserUUID})
			if err != nil {
				logger.Error("error in creating student: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student profile"})
				return
			}
		case "expert":
			err := expertRepo.Create(&models.Expert{UserID: user.UserUUID})
			if err != nil {
				logger.Error("error in creating expert: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expert profile"})
				return
			}
		}
	}

	// Generate access and refresh tokens
	accessToken, err := utils.GenerateAccessToken(user.UserUUID, user.Role)
	if err != nil {
		logger.Error("error in generating access token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.UserUUID, user.Role)
	if err != nil {
		logger.Error("error in generating refresh token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	setRefreshTokenCookie(c, refreshToken)

	// Respond
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"user": gin.H{
			"id":     user.ID,
			"name":   user.FullName,
			"email":  user.Email,
			"role":   user.Role,
			"avatar": user.Picture,
		},
	})
}

func UserSignIn(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		logger.Errorf("User not found: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if user.Password == nil || utils.VerifyPassword(*user.Password, req.Password) != nil {
		logger.Errorf("Invalid password for user: %s\n", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.UserUUID, user.Role)
	if err != nil {
		logger.Errorf("Error generating access token: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.UserUUID, user.Role)
	if err != nil {
		logger.Errorf("Error generating refresh token: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	setRefreshTokenCookie(c, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.FullName,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func RefreshSession(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		logger.Error("refresh token error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}

	claims, err := utils.ValidateRefreshToken(cookie)
	if err != nil {
		logger.Error("errr in verifying refresh token:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Generate new access token
	accessToken, _ := utils.GenerateAccessToken(claims.UserID, claims.Role)

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
