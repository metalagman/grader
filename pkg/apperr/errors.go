package apperr

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrNotFound          = fmt.Errorf("not found: %w", ErrInvalidInput)
	ErrConflict          = fmt.Errorf("conflict: %w", ErrInvalidInput)
	ErrSoftConflict      = fmt.Errorf("soft conflict: %w", ErrInvalidInput)
	ErrInsufficientFunds = fmt.Errorf("insufficient funds: %w", ErrInvalidInput)
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternal          = errors.New("internal")
)
