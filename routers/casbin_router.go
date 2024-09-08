package routers

import (
	casbinController "github.com/NubeDev/flexy/app/controllers/v1/casbin"
	"github.com/NubeDev/flexy/app/middleware"

	"github.com/gin-gonic/gin"
)

func InitCasbinRouter(Router *gin.RouterGroup) {
	casbin := Router.Group("/casbin").Use(
		middleware.TranslationHandler(),
		middleware.JWTHandler(),
		middleware.CasbinHandler())
	{
		casbin.GET("", casbinController.GetCasbinList)
		casbin.POST("", casbinController.CreateCasbin)
		casbin.PUT("/:id", casbinController.UpdateCasbin)
		casbin.DELETE("/:id", casbinController.DeleteCasbin)
	}
}
