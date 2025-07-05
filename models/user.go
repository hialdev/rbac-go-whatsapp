package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	Name       *string     `json:"name" gorm:"omitempty" validate:"required,min=2,max=20"`
	Username   *string     `json:"username" gorm:"omitempty;unique" validate:"required,min=4,max=12"`
	Password   *string     `json:"-"`
	Phone      string     `json:"phone" gorm:"unique" validate:"required,unique,min=4,max=14"`
	Image      *string     `json:"image" gorm:"text;omitempty"`
	VerifiedAt bool       `json:"verified_at,omitempty" validate:"omitempty,boolean"`
	RoleID     *uuid.UUID `json:"role_id,omitempty"`

	Role Role `json:"role,omitempty" gorm:"foreignKey:RoleID;constraint:SET NULL;"`
}

func (u *User) BeforeDelete(db *gorm.DB) (err error) {
	return
}
