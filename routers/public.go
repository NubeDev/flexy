package routers

import (
	indexController "github.com/NubeDev/flexy/app/controllers/v1/index"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitTestRouter(Router *gin.RouterGroup) {
	test := Router.Group("/test").Use(
		middleware.TranslationHandler(),
	)
	{
		test.GET("/ping", indexController.Ping)
	}
}
