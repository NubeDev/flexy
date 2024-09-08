package model

type Menu struct {
	UUID      string    `gorm:"primary_key" json:"uuid"`
	CreatedAt JSONTime  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt JSONTime  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *JSONTime `sql:"index" json:"deleted_at"`

	ParentId   int    `json:"parent_id" gorm:"type:int(11);"`
	Sort       int    `json:"sort" gorm:"type:int(4);"`
	MenuName   string `json:"menu_name" gorm:"type:varchar(11);comment:'Route name'"`
	Path       string `json:"path" gorm:"type:varchar(128);comment:'Route path'"`
	Paths      string `json:"paths" gorm:"type:varchar(128);"`
	Component  string `json:"component" gorm:"type:varchar(255);comment:'Component path'"`
	Title      string `json:"title" gorm:"type:varchar(64);comment:'Menu title'"`
	Icon       string `json:"icon" gorm:"type:varchar(128);"`
	MenuType   string `json:"menu_type" gorm:"type:varchar(1);"` // "M": Directory, "C": Menu, "F": Button
	Permission string `json:"permission" gorm:"type:varchar(32);"`
	Visible    string `json:"visible" gorm:"type:int(1);DEFAULT:0;"`
	IsFrame    string `json:"is_frame" gorm:"type:int(1);DEFAULT:0;"` // Whether it is an external link
	Params     string `json:"params" gorm:"-"`
	RoleId     int    `gorm:"-"`
	Children   []Menu `json:"children" gorm:"-"`
	IsSelect   bool   `json:"is_select" gorm:"-"`
	BaseModelNoId
}

func (Menu) TableName() string {
	return TablePrefix + "menu"
}
