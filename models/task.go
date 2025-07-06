package models

import (
	"github.com/google/uuid"
)

type Task struct {
	BaseModel
	TodoGroupID uuid.UUID `gorm:"type:uuid;not null;index" json:"todo_group_id" validate:"required"`
	Name        string    `gorm:"type:text;not null" json:"name" validate:"required"`
	Description string    `gorm:"type:text" json:"description"`
	AssignID    uuid.UUID `gorm:"type:uuid;index" json:"assign_id"`
	Status      string    `gorm:"type:text;default:'pending'" json:"status" validate:"oneof=wait process done"`
	
	TodoGroup   *TodoGroup `gorm:"foreignKey:TodoGroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"todo_group"`
	Assign      *User      `gorm:"foreignKey:AssignID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"assign"`
}
