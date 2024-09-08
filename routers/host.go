package routers

import (
	hostController "github.com/NubeDev/flexy/app/controllers/v1/host"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitHostRouter(Router *gin.RouterGroup) {
	endPoint := Router.Group("hosts").Use(middleware.TranslationHandler())
	if useAuth {
		endPoint.Use(
			middleware.JWTHandler(),
			middleware.CasbinHandler(),
		)
	}
	{
		endPoint.POST("", hostController.CreateHost)
		endPoint.GET("", hostController.GetHosts)
		endPoint.GET("/:uuid", hostController.GetHost)
		endPoint.PATCH("/:uuid", hostController.UpdateHost)
		endPoint.DELETE("/:uuid", hostController.DeleteHost)
	}
}
