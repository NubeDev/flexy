package hostController

import (
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
	"net/http"

	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/gin-gonic/gin"
)

func CreateHost(c *gin.Context) {
	g := common.Gin{C: c}
	// Bind payload to struct
	var body *hostService.Fields
	if err := c.ShouldBindJSON(&body); err != nil {
		g.Response(http.StatusBadRequest, code.InvalidParams, err.Error(), nil)
		return
	}
	// Validate bound struct parameters
	err, parameterErrorStr := common.CheckBindStructParameter(body, c)
	if err != nil {
		g.Response(http.StatusBadRequest, code.InvalidParams, parameterErrorStr, nil)
		return
	}

	resp, err := hostService.Create(body)
	g.Response(http.StatusOK, code.SUCCESS, "success", resp)
}
