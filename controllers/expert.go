package controllers

import (
	"interviewexcel-backend-go/config"
	"interviewexcel-backend-go/models"
	"net/http"
	"strings"
	"time"

	logger "interviewexcel-backend-go/pkg/errors"

	"github.com/gin-gonic/gin"
)

func GetExpertBookingsHandler(c *gin.Context) {
	var (
		availabilityRepo = models.InitAvailabilitySlotRepo(config.DB)
	)
	expertIDInterface, exists := c.Get("expert_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	expertID := expertIDInterface.(uint)

	slots, err := availabilityRepo.GetBookedSlotsByExpert(expertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, slots)
}

func GetExpertProfile(c *gin.Context) {
	var (
		expertRepo = models.InitExpertRepo(config.DB)
		userRepo   = models.InitUserRepo(config.DB)
	)

	user_id, exists := c.Get("user_uuid")
	if !exists {
		logger.Error("User doesn't exist")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uuid, ok := user_id.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user UUID"})
		return
	}

	expertResp, err := expertRepo.GetWithTx(config.DB, &models.Expert{UserID: uuid})
	if err != nil {
		logger.Error("error in getting expert: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Expert not found"})
		return
	}

	userResp, err := userRepo.GetByUUID(uuid)
	if err != nil {
		logger.Error("error in getting user: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	profile := ExpertProfile{
		UserID:   userResp.ID,
		UserUUID: userResp.UserUUID,
		FullName: userResp.FullName,
		Email:    userResp.Email,
		Picture:  userResp.Picture,
		Phone:    userResp.Phone,
		Role:     userResp.Role,

		Bio:                expertResp.Bio,
		DOB:                expertResp.DOB,
		Expertise:          expertResp.Expertise,
		ExperienceYears:    expertResp.ExperienceYears,
		ProfilePictureUrl:  expertResp.ProfilePictureUrl,
		Education:          expertResp.Education,
		City:               expertResp.City,
		Languages:          expertResp.Languages,
		IsAvailable:        expertResp.IsAvailable,
		FeesPerSession:     expertResp.FeesPerSession,
		Rating:             expertResp.Rating,        // if present
		TotalSessions:      expertResp.TotalSessions, // if present
		Specializations:    expertResp.Specializations,
		VerificationStatus: expertResp.VerificationStatus, // if present
		StudentMentored:    expertResp.StudentMentored,
	}

	c.JSON(http.StatusOK, profile)

}
func UpdateExpertProfile(c *gin.Context) {
	var request ExpertProfile

	userID, exists := c.Get("user_uuid")
	if !exists {
		logger.Error("user_uuid not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uuid, ok := userID.(string)
	if !ok {
		logger.Errorf("invalid user_uuid type: %T", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user UUID"})
		return
	}

	// Bind JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Errorf("failed to bind UpdateExpertProfile request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		logger.Errorf("failed to start transaction: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	expertRepo := models.InitExpertRepo(tx)
	userRepo := models.InitUserRepo(tx)

	expertRequest := &models.Expert{
		Expertise:         request.Expertise,
		ExperienceYears:   request.ExperienceYears,
		City:              request.City,
		DOB:               request.DOB,
		ProfilePictureUrl: request.ProfilePictureUrl,
		FeesPerSession:    request.FeesPerSession,
		Achievement:       request.Achievements,
	}

	if err := expertRepo.UpdateWithTx(tx, &models.Expert{UserID: uuid}, expertRequest); err != nil {
		tx.Rollback()
		logger.Errorf("failed to update expert profile (user_uuid=%s): %v", uuid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expert profile"})
		return
	}

	if err := userRepo.UpdateByUserUUID(uuid, &models.User{
		FullName: request.FullName,
		Phone:    request.Phone,
	}); err != nil {
		tx.Rollback()
		logger.Errorf("failed to update user profile (user_uuid=%s): %v", uuid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		logger.Errorf("failed to commit transaction (user_uuid=%s): %v", uuid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func GenerateWeeklyAvailability(c *gin.Context) {
	var (
		req              AvailabilityRequest
		availabilityRepo = models.InitAvailabilitySlotRepo(config.DB)
	)

	// Bind request
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find Monday of current week
	weekStart := time.Now()
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}

	var slots []models.AvailabilitySlot

	// Normalize requested days into a set for quick lookup
	daySet := make(map[string]bool)
	for _, d := range req.Days {
		daySet[strings.ToLower(d)] = true
	}
	// Iterate over the week
	for i := 0; i < 7; i++ {
		currentDay := weekStart.AddDate(0, 0, i)
		if !daySet[strings.ToLower(currentDay.Weekday().String())] {
			continue
		}
		logger.Info("here gef")
		// Parse clock times
		startClock, err := time.Parse("15:04", req.Start)
		if err != nil {
			logger.Error("invalid start time", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start time"})
			return
		}

		endClock, err := time.Parse("15:04", req.End)
		if err != nil {
			logger.Error("invalid end time", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end time"})
			return
		}

		// Merge with current day’s date
		start := time.Date(
			currentDay.Year(), currentDay.Month(), currentDay.Day(),
			startClock.Hour(), startClock.Minute(), 0, 0, time.Local,
		)
		end := time.Date(
			currentDay.Year(), currentDay.Month(), currentDay.Day(),
			endClock.Hour(), endClock.Minute(), 0, 0, time.Local,
		)
		if end.Before(start) || end.Equal(start) {
			end = end.AddDate(0, 0, 1)
		}

		if req.SlotSize <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slot size must be greater than 0"})
			return
		}

		duration := time.Minute * time.Duration(req.SlotSize)

		for t := start; t.Add(duration).Before(end) || t.Add(duration).Equal(end); t = t.Add(duration) {
			slot := models.AvailabilitySlot{
				ExpertID:  (req.ExpertID),
				Date:      time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local),
				StartTime: t,
				EndTime:   t.Add(duration),
				Status:    string(models.SlotAvailable),
			}
			slots = append(slots, slot)
		}

	}

	// Save slots
	if len(slots) == 0 {
		logger.Warn("no slots generated for given request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no slots generated"})
		return
	}

	if err := availabilityRepo.CreateAvailabilitySlot(slots); err != nil {
		logger.Error("error in generating slots", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"slots": slots})
}

type GetSlotsOfMerhantRequest struct {
	ExpertID string `json:"expert_id"`
}

func GetAllSlotsOfExpert(c *gin.Context) {
	var (
		availabilityRepo = models.InitAvailabilitySlotRepo(config.DB)
		// request          GetSlotsOfMerhantRequest
	)

	expertUUIDStr, ok := c.Get("user_uuid")
	if !ok {
		logger.Error("cannot find expert uuid")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "cannot find expert uuid",
		})
		return
	}
	uuid, ok := expertUUIDStr.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user UUID"})
		return
	}

	availableSlots, err := availabilityRepo.GetAllByExpert(uuid)
	if err != nil {
		logger.Error("Error in getting the available slots: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	c.JSON(http.StatusOK, availableSlots)
}

func CancelSlotOfExpert(c *gin.Context) {
	slotID := c.Param("slot_id")

	userID, exists := c.Get("user_uuid")
	if !exists {
		logger.Error("user_uuid missing in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	expertID := userID.(string)

	var slot models.AvailabilitySlot

	// Fetch slot
	if err := config.DB.
		Where("id = ? AND expert_id = ?", slotID, expertID).
		First(&slot).Error; err != nil {

		logger.Errorf("slot not found (slot_id=%s, expert_id=%s): %v", slotID, expertID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Slot not found"})
		return
	}

	// Business rules
	if slot.Status == string(models.SlotBooked) {
		logger.Warnf("attempt to cancel booked slot (slot_id=%d)", slot.ID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Booked slot cannot be cancelled",
		})
		return
	}

	if slot.Status == string(models.SlotCancelled) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Slot already cancelled",
		})
		return
	}

	// Cancel slot
	if err := config.DB.Model(&slot).
		Update("status", string(models.SlotCancelled)).Error; err != nil {

		logger.Errorf("failed to cancel slot (slot_id=%d): %v", slot.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel slot"})
		return
	}

	logger.Infof("slot cancelled successfully (slot_id=%d, expert_id=%s)", slot.ID, expertID)
	c.JSON(http.StatusOK, gin.H{"message": "Slot cancelled successfully"})
}

func GetExpertDashboard(c *gin.Context) {
	var (
		expertRepo       = models.InitExpertRepo(config.DB)
		userRepo         = models.InitUserRepo(config.DB)
		sessionRepo      = models.InitSessionRepo(config.DB)
		availabilityRepo = models.InitAvailabilitySlotRepo(config.DB)
		walletRepo       = models.InitWalletRepo(config.DB)
	)

	// 1. Get user UUID from auth context
	userIDInterface, exists := c.Get("user_uuid")
	if !exists {
		logger.Error("user_uuid not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uuid, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user UUID"})
		return
	}

	// 2. Fetch expert profile
	expert, err := expertRepo.GetWithTx(config.DB, &models.Expert{UserID: uuid})
	if err != nil {
		logger.Error("error fetching expert for dashboard: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Expert not found"})
		return
	}

	// 3. Fetch user details (for profile picture fallback)
	user, err := userRepo.GetByUUID(uuid)
	if err != nil {
		logger.Error("error fetching user for dashboard: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	profilePic := expert.ProfilePictureUrl
	if profilePic == "" {
		profilePic = user.Picture
	}

	// 4. Fetch wallet for earnings
	var earningsInPaise int64
	wallet, err := walletRepo.GetByUserUUID(uuid)
	if err == nil && wallet != nil {
		earningsInPaise = wallet.BalanceInPaise
	}

	// 5. Fetch upcoming sessions
	sessions, err := sessionRepo.GetUpcomingForUser(uuid)
	if err != nil {
		logger.Error("error fetching upcoming sessions: ", err)
		sessions = []models.Session{}
	}

	// Enrich sessions with student names
	var upcomingSessions []ExpertSessionResponse
	for _, session := range sessions {
		studentName := "Unknown Student"
		studentUser, err := userRepo.GetByUUID(session.StudentUUID)
		if err == nil && studentUser != nil {
			studentName = studentUser.FullName
		}

		upcomingSessions = append(upcomingSessions, ExpertSessionResponse{
			ID:          session.ID,
			SessionUUID: session.SessionUUID,
			StudentUUID: session.StudentUUID,
			StudentName: studentName,
			StartTime:   session.StartTime,
			EndTime:     session.EndTime,
			MeetLink:    session.MeetLink,
			Status:      session.Status,
		})
	}

	if upcomingSessions == nil {
		upcomingSessions = []ExpertSessionResponse{}
	}

	// 6. Fetch slot overview counts
	availableSlots, err := availabilityRepo.CountAvailableSlotsByExpert(uuid)
	if err != nil {
		logger.Error("error counting available slots: ", err)
		availableSlots = 0
	}

	bookedSlots, err := availabilityRepo.CountBookedSlotsByExpertUUID(uuid)
	if err != nil {
		logger.Error("error counting booked slots: ", err)
		bookedSlots = 0
	}

	// 7. Assemble response
	response := ExpertDashboardResponse{
		Expert: DashboardExpertInfo{
			FullName:           user.FullName,
			VerificationStatus: expert.VerificationStatus,
			IsAvailable:        expert.IsAvailable,
			ProfilePictureUrl:  profilePic,
		},
		Stats: DashboardStats{
			TotalSessions:    expert.TotalSessions,
			StudentsMentored: expert.StudentMentored,
			Rating:           expert.Rating,
			Earnings:         earningsInPaise,
		},
		UpcomingSessions: upcomingSessions,
		SlotOverview: DashboardSlotOverview{
			AvailableSlots: availableSlots,
			BookedSlots:    bookedSlots,
			SessionFee:     expert.FeesPerSession,
		},
	}

	c.JSON(http.StatusOK, response)
}
