package controllers

import (
	"encoding/json"
	"fmt"
	"interviewexcel-backend-go/config"
	"io"
	"net/http"

	// "yourapp/internal/oauth"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type GoogleUser struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func GoogleLoginHandler(c *gin.Context) {
	fmt.Print("i am here")
	url := config.GoogleConfig().AuthCodeURL("random-state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code in request"})
		return
	}

	token, err := config.GoogleConfig().Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed"})
		return
	}

	client := config.GoogleConfig().Client(c, token)
	res, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	var user GoogleUser
	if err := json.Unmarshal(body, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	// 💡 You can now identify user type (student or expert) and create or update DB record accordingly

	c.JSON(http.StatusOK, gin.H{"user": user})
}
