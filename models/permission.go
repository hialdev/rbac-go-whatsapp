package models

type Permission struct {
	BaseModel
	Name        string  `gorm:"type:varchar(100)" json:"name" validate:"required,min=3"`
	Description *string `gorm:"type:text" json:"description" validate:"omitempty"`

	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}
