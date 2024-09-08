package model

import (
	"errors"
	"gorm.io/gorm"
)

type Role struct {
	RoleId    uint      `gorm:"primary_key" json:"role_id"` // Role Code
	CreatedAt JSONTime  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt JSONTime  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *JSONTime `sql:"index" json:"deleted_at"`
	RoleName  string    `gorm:"type:varchar(128);" json:"role_name"` // Role Name
	IsAdmin   bool      `gorm:"type:int(1);DEFAULT:0;NOT NULL;" json:"is_admin"`
	Status    int       `gorm:"type:int(1);DEFAULT:0;NOT NULL;" json:"status"`
	RoleKey   string    `gorm:"type:varchar(128);unique;" json:"role_key"` // Role Code
	RoleSort  int       `gorm:"type:int(4);" json:"role_sort"`             // Role Sort Order
	Remark    string    `gorm:"type:varchar(255);" json:"remark"`          // Remarks
	Params    string    `gorm:"-" json:"params"`
	MenuIds   []int     `gorm:"-" json:"menu_ids"`
}

func (Role) TableName() string {
	return TablePrefix + "role"
}

func CreateRole(role Role) error {
	res := db.Create(&role)
	if err := res.Error; err != nil {
		return err
	}
	return nil
}

func GetRoles(pageNum int, pageSize int, where map[string]interface{}) ([]*Role, error) {
	var m []*Role
	db, _ := BuildCondition(db, where)
	err := db.Select("*").Offset(pageNum).Limit(pageSize).Find(&m).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return m, nil
}
