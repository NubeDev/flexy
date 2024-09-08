package routers

import (
	rqlController "github.com/NubeDev/flexy/app/controllers/v1/rql"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitRQLRouter(Router *gin.RouterGroup) {
	endPoint := Router.Group("rql").Use(middleware.TranslationHandler())
	if useAuth {
		endPoint.Use(
			middleware.JWTHandler(),
			middleware.CasbinHandler(),
		)
	}
	{
		endPoint.POST("/run", rqlController.RQL)
	}
}
