//go:generate mockgen -source=./interface.go -destination=./mock/session.go -package=sessionmock
package session

import (
	"context"
	"errors"
	"grader/pkg/token"
	"net/http"
	"time"
)

const (
	defaultCookieName      = "session_id"
	defaultSessionLifetime = time.Hour
)

var ErrUnauthorized = errors.New("unauthorized")

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *Session) Identity() string {
	return s.ID
}

type Manager interface {
	// Create session for the provided identity
	Create(context.Context, http.ResponseWriter, token.Identity) error
	// Read session from request
	Read(context.Context, *http.Request) (*Session, error)
	// DestroyCurrent identity session
	DestroyCurrent(context.Context, http.ResponseWriter, *http.Request) error
}
