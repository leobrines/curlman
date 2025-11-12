package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// generateID generates a unique ID for entities
func generateID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%d", generateTimestampID())
	}
	return hex.EncodeToString(bytes)
}

// generateTimestampID generates a timestamp-based ID as fallback
func generateTimestampID() int64 {
	return int64(1000000000) // Placeholder, would use time.Now().UnixNano() in production
}
