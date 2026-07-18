package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"interviewexcel-backend-go/config"
	"os"
)

// Student clicks "Book"
//    ↓
// Backend creates Razorpay Order
//    ↓
// Frontend opens Razorpay Checkout
//    ↓
// Payment success
//    ↓
// Frontend sends payment_id + order_id + signature
//    ↓
// Backend verifies signature
//    ↓
// BEGIN TX
//    ├─ Lock slot
//    ├─ Create session
//    ├─ Mark slot booked
// COMMIT

type RazorpayOrderResponse struct {
	OrderID  string `json:"order_id"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
	Key      string `json:"key"`
}

func CreateRazorpayOrder(slotID uint, amountInPaise int) (*RazorpayOrderResponse, error) {

	data := map[string]interface{}{
		"amount":   amountInPaise,
		"currency": "INR",
		"receipt":  fmt.Sprintf("slot_%d", slotID),
	}

	order, err := config.RazorpayClient.Order.Create(data, nil)
	if err != nil {
		return nil, err
	}

	resp := &RazorpayOrderResponse{
		OrderID:  order["id"].(string),
		Amount:   amountInPaise,
		Currency: "INR",
		Key:      os.Getenv("RAZORPAY_KEY"),
	}

	return resp, nil
}

func VerifyRazorpaySignature(orderID string, paymentID string, signature string) bool {

	secret := os.Getenv("RAZORPAY_SECRET")

	data := orderID + "|" + paymentID

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))

	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return expectedSignature == signature
}
