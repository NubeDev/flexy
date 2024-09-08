package model

import (
	"errors"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type CasbinRuleM struct {
	BaseModel
	Ptype string `gorm:"type:varchar(300);"`
	V0    string `json:"v0" gorm:"type:varchar(100);uniqueIndex:unique_index"`
	V1    string `json:"v1" gorm:"type:varchar(100);uniqueIndex:unique_index"`
	V2    string `json:"v2" gorm:"type:varchar(100);uniqueIndex:unique_index"`
	V3    string `json:"v3" gorm:"type:varchar(100);uniqueIndex:unique_index"`
	V4    string `json:"v4" gorm:"type:varchar(100);uniqueIndex:unique_index"`
	V5    string `json:"v5" gorm:"type:varchar(100);uniqueIndex:unique_index"`
}

func (CasbinRuleM) TableName() string {
	return "casbin_rule"
}

func SetupCasbin() *casbin.SyncedEnforcer {
	a, _ := gormadapter.NewAdapterByDBWithCustomTable(db, &CasbinRuleM{})
	e, _ := casbin.NewSyncedEnforcer("config/rbac_model.conf", a)

	// Refresh every 12 hours.
	e.StartAutoLoadPolicy(12 * time.Hour)

	return e

}

func CreatCasbin(casbin CasbinRuleM) error {
	res := db.Create(&casbin)
	if err := res.Error; err != nil {
		return err
	}
	return nil
}

func GetCasbinRuleList(pageNum int, pageSize int, where map[string]interface{}) ([]*CasbinRuleM, error) {
	var m []*CasbinRuleM
	db, _ := BuildCondition(db, where)
	err := db.Select("*").Offset(pageNum).Limit(pageSize).Find(&m).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return m, nil
}
