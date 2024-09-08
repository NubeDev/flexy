package sysController

import (
	"net/http"

	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/gin-gonic/gin"
)

var Routers gin.RoutesInfo

func GetRouterList(c *gin.Context) {
	appG := common.Gin{C: c}

	type Router struct {
		Path   string `json:"path"`
		Method string `json:"method"`
	}

	data := make([]Router, 0)

	routers := Routers

	for index := range routers {
		var router Router
		router.Method = Routers[index].Method
		router.Path = Routers[index].Path
		data = append(data, router)
	}

	appG.Response(http.StatusOK, code.SUCCESS, "Successfully retrieved existing route list", data)
}
