package routers

import (
	sysController "github.com/NubeDev/flexy/app/controllers/v1/sys"
	"github.com/NubeDev/flexy/utils/setting"

	"github.com/gin-gonic/gin"
)

var useAuth = setting.ServerSetting.UseAuth

func InitRouter(r *gin.Engine) *gin.Engine {
	v1 := r.Group("/v1/api")
	{
		InitUserRouter(v1)
		InitRoleRouter(v1)
		InitCasbinRouter(v1)
		InitSysRouter(v1)
		InitTestRouter(v1)
		InitReportRouter(v1)
		InitHostRouter(v1)
		InitRQLRouter(v1)
	}

	// Route list
	sysController.Routers = r.Routes()

	return r
}
