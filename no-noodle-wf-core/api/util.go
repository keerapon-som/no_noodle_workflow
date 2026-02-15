package api

import (
	"github.com/google/uuid"
)

func generateWorkflowID() string {
	return uuid.New().String()
}

func generateSessionKey() string {
	return uuid.New().String()
}
