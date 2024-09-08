package routers

import (
	casbinController "github.com/NubeDev/flexy/app/controllers/v1/casbin"
	"github.com/NubeDev/flexy/app/middleware"

	"github.com/gin-gonic/gin"
)

func InitCasbinRouter(router *gin.RouterGroup) {
	endPoint := router.Group("/api/casbin").Use(middleware.TranslationHandler())
	if useAuth {
		endPoint.Use(
			middleware.JWTHandler(),
			middleware.CasbinHandler(),
		)
	}
	{
		endPoint.GET("", casbinController.GetCasbinList)
		endPoint.POST("", casbinController.CreateCasbin)
		endPoint.PUT("/:id", casbinController.UpdateCasbin)
		endPoint.DELETE("/:id", casbinController.DeleteCasbin)
	}
}
