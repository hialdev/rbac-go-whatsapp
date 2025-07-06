package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatSummary struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	Summary   string    `gorm:"type:text;not null" json:"summary" validate:"required"`
	StartTime time.Time `gorm:"not null" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`
	
	User      *User      `gorm:"foreignKey:UserID" json:"user"`
}
