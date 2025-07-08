package models

import "github.com/google/uuid"

type Notification struct {
	BaseModel
	Title       string    `gorm:"type:text;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Issue       string    `gorm:"type:text;not null" json:"issue" validate:"oneof=discussion assign done"`
	TodoGroupID uuid.UUID `gorm:"type:uuid;index" json:"todo_group_id"`
	TaskID      *uuid.UUID `gorm:"type:uuid;index" json:"task_id,omitempty"`
	IsRead      bool      `gorm:"default:false" json:"is_read"`

	TodoGroup *TodoGroup `gorm:"foreignKey:TodoGroupID" json:"todo_group,omitempty"`
	Task      *Task      `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}
