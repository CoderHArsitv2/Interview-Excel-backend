package utils

import (
	"interviewexcel-backend-go/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GenerateWeeklyAvailability(c *gin.Context) {
	var input struct {
		ExpertID string `json:"expert_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slots := GenerateWeeklySlots(input.ExpertID)

	if err := config.DB.Create(&slots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create availability slots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Availability slots created successfully"})
}
