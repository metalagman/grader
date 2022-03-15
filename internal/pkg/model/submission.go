package model

import (
	"github.com/google/uuid"
	"time"
)

type Submission struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       uuid.UUID `json:"user_id"`
	AssessmentID uuid.UUID `json:"assessment_id"`
	FileName     string    `json:"file_name"`
	FileURL      string    `json:"file_url"`
	ExternalID   string    `json:"external_id"`
	ResultDate   time.Time `json:"result_date"`
	ResultPass   bool      `json:"result_pass"`
	ResultText   string    `json:"result_text"`
}
