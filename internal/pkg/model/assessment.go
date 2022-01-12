package model

import (
	"github.com/google/uuid"
	"time"
)

type Assessment struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	PartID         string    `json:"part_id"`
	ContainerImage string    `json:"container_image"`
	Summary        string    `json:"summary"`
	FileName       string    `json:"file_name"`
}
