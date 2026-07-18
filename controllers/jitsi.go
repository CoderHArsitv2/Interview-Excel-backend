package controllers

import (
	"crypto/sha256"
	"fmt"
	"time"
)

func generateJitsiMeetLink(
	expertUUID string,
	studentUUID string,
	start time.Time,
) string {

	raw := fmt.Sprintf(
		"interviewexcel-%s-%s-%d",
		expertUUID,
		studentUUID,
		start.Unix(),
	)

	hash := sha256.Sum256([]byte(raw))

	return fmt.Sprintf(
		"https://meet.jit.si/ie-%x",
		hash[:10],
	)
}
