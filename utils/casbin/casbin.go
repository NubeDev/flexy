package casbin

import (
	model "github.com/NubeDev/flexy/app/models"

	"github.com/casbin/casbin/v2"
)

var (
	CasbinEnforcer *casbin.SyncedEnforcer
)

func SetupCasbin() *casbin.SyncedEnforcer {
	CasbinEnforcer = model.SetupCasbin()
	return CasbinEnforcer
}
