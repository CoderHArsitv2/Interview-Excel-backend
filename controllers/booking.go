package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	logger "interviewexcel-backend-go/pkg/errors"
)

// Booking flow (what the function must guarantee)
// Atomic requirements

// Slot must exist

// Slot must be available

// Slot must be booked once

// Google Meet link must be created

// If any step fails → rollback

// High-level flow
// BEGIN TX
//   ├─ Lock availability slot
//   ├─ Validate slot availability
//   ├─ Create Google Meet link
//   ├─ Create booking record
//   ├─ Mark slot as booked
// COMMIT

type BookSlotRequest struct {
	SlotID        uint `json:"slot_id" binding:"required"`
	AmountInPaise int  `json:"amount_in_paise" binding:"required"`

	// In future: StudentID uint `json:"student_id"`
}

func InitiateBookingHandler(c *gin.Context) {
	var (
		req BookSlotRequest
	)
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logger.Error("error in binding booking request: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := CreateRazorpayOrder(req.SlotID, req.AmountInPaise)
	if err != nil {
		logger.Error("error in creating razorpay order: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment order"})
		return
	}
	c.JSON(http.StatusOK, order)
}

type ConfirmPaymentRequest struct {
	SlotID            uint   `json:"slot_id" binding:"required"`
	RazorpayOrderID   string `json:"razorpay_order_id" binding:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" binding:"required"`
	RazorpaySignature string `json:"razorpay_signature" binding:"required"`
}

type ConfirmPaymentResponse struct {
	SessionID uint `json:"session_id"`
}

func ConfirmPaymentHandler(c *gin.Context) {
	var (
		req ConfirmPaymentRequest
	)
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logger.Error("error in binding payment confirmation request: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok := VerifyRazorpaySignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature)
	if !ok {
		logger.Error("error in verifying razorpay signature")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment signature"})
		return
	}

	err = BookExpertSlot(c, req.SlotID)
	if err != nil {
		logger.Error("error in booking slot after payment: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to book slot"})
		return
	}

	// Sending session details to student in response
	c.JSON(http.StatusOK,
		ConfirmPaymentResponse{
			SessionID: req.SlotID,
		},
	)

}
