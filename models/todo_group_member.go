package models

import (
	"time"

	"github.com/google/uuid"
)

type TodoGroupMember struct {
	BaseModel
	TodoGroupID uuid.UUID  `gorm:"type:uuid;not null;index" json:"todo_group_id" validate:"required"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	JoinedAt    time.Time  `gorm:"autoCreateTime" json:"joined_at"`
	
	TodoGroup   *TodoGroup  `gorm:"foreignKey:TodoGroupID" json:"todo_group"`
	User        *User       `gorm:"foreignKey:UserID" json:"user"`
}
