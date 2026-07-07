package chi_types

import (
	"time"

	"github.com/google/uuid"
)

type ModelBase struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type ModelBaseWithArchive struct {
	ModelBase
	DeletedAt *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}
