package utils

import (
	"interviewexcel-backend-go/models"
	"time"
)

func GenerateWeeklySlots(expertID string) []models.AvailabilitySlot {
	var slots []models.AvailabilitySlot
	now := time.Now()

	startHour := 9 // 9 AM
	endHour := 21  // 9 PM (exclusive, generates 12 slots per day)

	for day := 0; day < 7; day++ {
		currentDate := now.AddDate(0, 0, day)
		for hour := startHour; hour < endHour; hour++ {
			startTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), hour, 0, 0, 0, time.UTC)
			endTime := startTime.Add(time.Hour)

			slots = append(slots, models.AvailabilitySlot{
				ExpertID:  expertID,
				Date:      currentDate,
				StartTime: startTime,
				EndTime:   endTime,
				Status:    string(models.SlotAvailable),
			})
		}
	}
	return slots
}
