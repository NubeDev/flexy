package indexController

import (
	"net/http"

	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	appG := common.Gin{C: c}
	appG.Response(http.StatusOK, code.SUCCESS, "pong", nil)
}
