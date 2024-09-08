package routers

import (
	sysController "github.com/NubeDev/flexy/app/controllers/v1/sys"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitSysRouter(router *gin.RouterGroup) {
	endPoint := router.Group("system").Use(middleware.TranslationHandler())
	if useAuth {
		endPoint.Use(
			middleware.JWTHandler(),
			middleware.CasbinHandler(),
		)
	}
	{
		endPoint.GET("/router", sysController.GetRouterList) // Route list
	}
}
