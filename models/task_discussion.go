package models

import (
	"github.com/google/uuid"
)

type TaskDiscussion struct {
	BaseModel
	TaskID  uuid.UUID `gorm:"type:uuid;not null;index" json:"task_id" validate:"required"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	Message string    `gorm:"type:text;not null" json:"message" validate:"required"`

	Task        *Task            `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	User        *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
