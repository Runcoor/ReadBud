package domain

import (
	"time"

	"gorm.io/gorm"

	"readbud/internal/pkg/utils"
)

// BaseModel contains shared audit fields for all domain entities.
// Matches spec Section 11 unified audit fields.
type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID  string         `gorm:"column:public_id;type:varchar(26);uniqueIndex;not null" json:"public_id"`
	CreatedAt time.Time      `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index" json:"deleted_at,omitempty"`
	CreatedBy *int64         `gorm:"type:bigint" json:"created_by,omitempty"`
	UpdatedBy *int64         `gorm:"type:bigint" json:"updated_by,omitempty"`
}

// BeforeCreate generates a ULID for public_id if not already set.
func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.PublicID == "" {
		b.PublicID = utils.NewULID()
	}
	return nil
}

// Status constants used across multiple domains.
const (
	StatusActive   int16 = 1
	StatusInactive int16 = 0
)
