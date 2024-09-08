package routers

import (
	roleController "github.com/NubeDev/flexy/app/controllers/v1/role"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitRoleRouter(Router *gin.RouterGroup) {
	role := Router.Group("/role").Use(
		middleware.TranslationHandler(),
		middleware.JWTHandler(),
		middleware.CasbinHandler())
	{
		role.GET("", roleController.GetRoles)
		role.POST("", roleController.CreateRole)
		role.PUT("/:role_id", roleController.UpdateRole)
		role.DELETE("/:role_id", roleController.DeleteRole)
	}
}
