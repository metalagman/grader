package runner

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type SubmissionResult struct {
	TaskID uuid.UUID `json:"task_id"`
	Pass   bool      `json:"pass"`
	Text   string    `json:"text"`
}

// sendResult to callback URL
func sendResult(ctx context.Context, URL string, result SubmissionResult) error {
	client := resty.New()

	_, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(result).
		Post(URL)
	if err != nil {
		return fmt.Errorf("result postback: %w", err)
	}

	return nil
}
