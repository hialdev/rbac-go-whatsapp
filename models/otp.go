package models

import (
	"time"
)

type Otp struct {
	BaseModel
	Phone     string    `json:"phone" gorm:"required,min=6,max=14"`
	Code      string    `json:"code" gorm:"unique" validate:"required,len=6"`
	ExpiredAt time.Time `json:"expired_at"`
	Purpose   string    `json:"purpose" validate:"oneof=register changes verify"`
}

func (Otp) TableName() string {
    return "otp"
}