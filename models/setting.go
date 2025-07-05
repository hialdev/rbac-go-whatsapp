package models

type Setting struct {
	BaseModel
	Name        string  `json:"name" gorm:"type:varchar(300)"`
	Description *string `json:"description,omitempty" gorm:"type:text;omitempty"`
	SetKey      string  `json:"set_key" gorm:"unique;type:varchar(100)" validate:"required,unique"`
	SetGroupKey string  `json:"set_group_key" gorm:"type:varchar(100);" validate:"required"`
	SetValue    *string `json:"set_value,omitempty" gorm:"type:text;omitempty" validate:"omitempty"`
	SetType     string  `json:"set_type" gorm:"type:varchar(100);default:'text'" validate:"oneof=text number checkbox radio select selects file image files images richtext markdown"`
	IsUrgent    bool    `json:"is_urgent" gorm:"type:bool;default:false"`
}

func (Setting) TableName() string {
    return "settings"
}
