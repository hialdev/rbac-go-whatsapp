package models

type TodoGroup struct {
	BaseModel
	Name        string `gorm:"type:text;not null" json:"name" validate:"required"`
	Description string `gorm:"type:text" json:"description"`

	Tasks   []Task            `gorm:"foreignKey:TodoGroupID" json:"tasks"`
	Members []TodoGroupMember `gorm:"foreignKey:TodoGroupID;references:ID" json:"members"`
}
