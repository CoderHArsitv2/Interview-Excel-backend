package controllers

import (
	"context"
	"errors"
	"interviewexcel-backend-go/config"
	"interviewexcel-backend-go/models"
	"time"

	logger "interviewexcel-backend-go/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/calendar/v3"
	"gorm.io/gorm/clause"
)

func CreateGoogleMeetLink(
	ctx context.Context,
	srv *calendar.Service,
	start, end time.Time,
) (string, error) {

	event := &calendar.Event{
		Summary: "InterviewExcel Expert Session",
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: "Asia/Kolkata",
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
			TimeZone: "Asia/Kolkata",
		},
		ConferenceData: &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId: uuid.New().String(),
			},
		},
	}

	createdEvent, err := srv.Events.
		Insert("primary", event).
		ConferenceDataVersion(1).
		Do()

	if err != nil {
		logger.Error("error in creating google meet link: ", err)
		return "", err
	}

	logger.Infof("event id=%s hangout=%s conference=%+v",
		createdEvent.Id,
		createdEvent.HangoutLink,
		createdEvent.ConferenceData,
	)
	return createdEvent.HangoutLink, nil
}

func BookExpertSlot(c *gin.Context, slotID uint) error {

	var (
		tx                   = config.DB.Begin()
		sessionRepo          = models.InitSessionRepo(tx)
		AvailabilitySlotRepo = models.InitAvailabilitySlotRepo(tx)
		walletRepo           = models.InitWalletRepo(tx)
		wtRepo               = models.InitWalletTransactionRepo(tx)
		expertRepo           = models.InitExpertRepo(tx)
	)
	studentUUID := c.GetString("user_uuid")

	if tx.Error != nil {
		logger.Error("error in starting transaction: ", tx.Error)
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var slot models.AvailabilitySlot
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND status = ?", slotID, models.SlotAvailable).
		First(&slot).Error

	if err != nil {
		tx.Rollback()
		logger.Error("slot not available: ", err)
		return errors.New("slot not available")
	}

	//FIXME: Temporarily disabling Google Meet link creation
	//FEATURE: ADDING PROVIDERS FOR MEET LINK CREATION

	// // 2️⃣ Create Google Calendar service
	// calService, err := calendar.NewService(
	// 	context.Background(),
	// 	option.WithCredentialsFile(os.Getenv("GOOGLE_CREDENTIALS_JSON")),
	// )
	// if err != nil {
	// 	tx.Rollback()
	// 	logger.Error("error in creating calendar service: ", err)
	// 	return err
	// }

	// 3️⃣ Create Meet link
	// meetLink, err := CreateGoogleMeetLink(
	// 	context.Background(),
	// 	calService,
	// 	slot.StartTime,
	// 	slot.EndTime,
	// )
	// if err != nil {
	// 	tx.Rollback()
	// 	logger.Error("error in creating google meet link: ", err)
	// 	return err
	// }

	session := &models.Session{
		SessionUUID: uuid.New().String(),
		ExpertUUID:  slot.ExpertID,
		StudentUUID: studentUUID,
		SlotID:      slot.ID,
		StartTime:   slot.StartTime,
		EndTime:     slot.EndTime,
		Status:      "scheduled",
	}

	session.MeetLink = generateJitsiMeetLink(
		session.ExpertUUID,
		session.StudentUUID,
		session.StartTime,
	)

	if err := sessionRepo.Create(session); err != nil {
		logger.Error("error in creating session: ", err)
		tx.Rollback()
		return err
	}

	// 5️⃣ Mark slot as booked
	err = AvailabilitySlotRepo.UpdateWithTx(
		tx,
		&models.AvailabilitySlot{
			Status: string(models.SlotBooked),
		}, &models.AvailabilitySlot{
			ID: slot.ID,
		})
	if err != nil {
		logger.Error("error in marking slot as booked: ", err)
		tx.Rollback()
		return err
	}

	expertDetails, err := expertRepo.GetWithTx(tx, &models.Expert{
		UserID: slot.ExpertID,
	})
	if err != nil {
		logger.Error("error in fetching expert details: ", err)
		tx.Rollback()
		return err
	}

	//Crediting to expert wallet next step
	wallet, err := walletRepo.GetByUserUUID(slot.ExpertID)
	if err != nil {
		// If wallet doesn't exist, create it
		wallet = &models.Wallet{
			UserUUID:       slot.ExpertID,
			BalanceInPaise: 0,
		}
		if err := walletRepo.Create(wallet); err != nil {
			logger.Error("error in creating expert wallet: ", err)
			tx.Rollback()
			return err
		}
	}

	// Update balance
	newBalance := wallet.BalanceInPaise + int64(expertDetails.FeesPerSession)
	if err := walletRepo.UpdateBalance(slot.ExpertID, newBalance); err != nil {
		logger.Error("error in updating expert wallet balance: ", err)
		tx.Rollback()
		return err
	}

	// Create wallet transaction
	err = wtRepo.Create(tx, &models.WalletTransaction{
		WalletID:      wallet.ID,
		AmountInPaise: int64(expertDetails.FeesPerSession),
		Type:          "credit",
		Source:        "session",
		ReferenceID:   session.SessionUUID,
		Description:   "Payment for session booking",
	})
	if err != nil {
		logger.Error("error in creating wallet transaction: ", err)
		tx.Rollback()
		return err
	}

	// All good, commit tx
	return tx.Commit().Error
}
