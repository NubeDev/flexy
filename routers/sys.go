package routers

import (
	sysController "github.com/NubeDev/flexy/app/controllers/v1/sys"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitSysRouter(Router *gin.RouterGroup) {
	sys := Router.Group("/sys").Use(
		middleware.TranslationHandler(),
		middleware.JWTHandler(),
		middleware.CasbinHandler())
	{
		sys.GET("/router", sysController.GetRouterList) // Route list
	}
}
