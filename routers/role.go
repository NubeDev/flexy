package routers

import (
	roleController "github.com/NubeDev/flexy/app/controllers/v1/role"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitRoleRouter(router *gin.RouterGroup) {
	endPoint := router.Group("/role").Use(middleware.TranslationHandler())
	if !useAuth {
		endPoint.Use(
			middleware.JWTHandler(),
			middleware.CasbinHandler(),
		)
	}

	{
		endPoint.GET("", roleController.GetRoles)
		endPoint.POST("", roleController.CreateRole)
		endPoint.PUT("/:role_id", roleController.UpdateRole)
		endPoint.DELETE("/:role_id", roleController.DeleteRole)
	}
}
