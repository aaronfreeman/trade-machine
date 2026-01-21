package models

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents an encrypted API key stored in the database
type APIKey struct {
	ID                 uuid.UUID `json:"id"`
	ServiceName        string    `json:"service_name"`
	APIKeyEncrypted    []byte    `json:"-"` // Never expose encrypted data in JSON
	APISecretEncrypted []byte    `json:"-"` // Never expose encrypted data in JSON
	BaseURL            string    `json:"base_url,omitempty"`
	Region             string    `json:"region,omitempty"`
	ModelID            string    `json:"model_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
