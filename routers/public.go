package routers

import (
	indexController "github.com/NubeDev/flexy/app/controllers/v1/public"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitTestRouter(router *gin.RouterGroup) {
	test := router.Group("public").Use(
		middleware.TranslationHandler(),
	)
	{
		test.GET("/ping", indexController.Ping)
	}
}
