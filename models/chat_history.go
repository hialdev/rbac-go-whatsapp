package models

import (
	"github.com/google/uuid"
)

type ChatHistory struct {
	BaseModel
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	Role    string    `gorm:"type:text;not null" json:"role" validate:"required,oneof=user assistant system"`
	Message string    `gorm:"type:text;not null" json:"message" validate:"required"`
	Meta    string    `gorm:"type:json" json:"meta"`
	
	User    *User      `gorm:"foreignKey:UserID" json:"user"` // Relasi ke tabel User
}
