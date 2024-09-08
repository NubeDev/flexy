package model

import "log"

type JwtBlacklist struct {
	UUID      string    `gorm:"primary_key" json:"uuid"`
	CreatedAt JSONTime  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt JSONTime  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *JSONTime `sql:"index" json:"deleted_at"`
	UserID    uint      `json:"user_id"`
	Jwt       string    `gorm:"type:text"`
}

func (JwtBlacklist) TableName() string {
	return TablePrefix + "jwt_blacklist"
}

func CreateBlockList(userId uint, jwt string) error {
	m := JwtBlacklist{UserID: userId, Jwt: jwt}
	res := db.Create(&m)
	if err := res.Error; err != nil {
		log.Println(err)
		return err
	}
	return nil
}
