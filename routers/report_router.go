package routers

import (
	reportController "github.com/NubeDev/flexy/app/controllers/v1/report"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitReportRouter(Router *gin.RouterGroup) {
	report := Router.Group("/report").Use(
		middleware.TranslationHandler())
	{
		report.POST("", reportController.Report)
	}
}
