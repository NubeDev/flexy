package routers

import (
	reportController "github.com/NubeDev/flexy/app/controllers/v1/report"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitReportRouter(router *gin.RouterGroup) {
	endPoint := router.Group("/api/reports").Use(middleware.TranslationHandler())
	if useAuth {
		endPoint.Use(
			middleware.JWTHandler(),
			middleware.CasbinHandler(),
		)
	}
	{
		endPoint.POST("", reportController.Report)
	}
}
