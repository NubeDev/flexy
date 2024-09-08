package routers

import (
	sysController "github.com/NubeDev/flexy/app/controllers/v1/sys"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) *gin.Engine {
	v1 := r.Group("/v1/api")
	{
		InitUserRouter(v1) // User management
		InitRoleRouter(v1) // Roles
		InitCasbinRouter(v1)
		InitSysRouter(v1)    // System settings
		InitTestRouter(v1)   // Test routes
		InitReportRouter(v1) // Reporting
	}

	// Route list
	sysController.Routers = r.Routes()

	return r
}
