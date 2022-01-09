//go:generate mockgen -source=./interface.go -destination=./mock/storage.go -package=storagemock
package storage

import (
	"context"
	"github.com/google/uuid"
	"grader/internal/app/panel/model"
)

type UserRepository interface {
	// Create a new model.User
	Create(ctx context.Context, m *model.User) (*model.User, error)
	// ReadByNameAndPassword instance of model.User
	ReadByNameAndPassword(ctx context.Context, name string, password string) (*model.User, error)
	// Read instance of model.User
	Read(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type AssessmentRepository interface {
	// Create a new model.Assessment
	Create(ctx context.Context, m *model.Assessment) (*model.Assessment, error)
	// All instances of model.Assessment
	All(ctx context.Context) ([]*model.Assessment, error)
	// Read instance of model.Assessment
	Read(ctx context.Context, id uuid.UUID) (*model.Assessment, error)
}
