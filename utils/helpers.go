package utils

import (
	"strings"

	"github.com/google/uuid"
)

func GenerateUserUUID(role string) string {
	prefix := ""
	switch strings.ToLower(role) {
	case "student":
		prefix = "st_"
	case "expert":
		prefix = "ex_"
	default:
		prefix = "usr_"
	}

	// Take first 8 chars of UUID for shorter ID
	shortID := uuid.New().String()[:8]

	return prefix + shortID
}
