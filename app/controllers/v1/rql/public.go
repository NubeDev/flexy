package rqlController

import (
	"github.com/NubeDev/flexy/app/services/rqlservice"
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/helpers"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Body struct {
	Script string `json:"script"`
}

func RQL(c *gin.Context) {
	g := common.Gin{C: c}
	var body *Body
	if err := c.ShouldBindJSON(&body); err != nil {
		g.Response(http.StatusBadRequest, code.InvalidParams, err.Error(), nil)
		return
	}
	destroy, err := rqlservice.RQL().RunAndDestroy(helpers.UUID(), body.Script, nil)
	if err != nil {
		g.Response(http.StatusOK, code.ERROR, "RQL error", err)
		return
	}
	g.Response(http.StatusOK, code.SUCCESS, "RQL ok", destroy)
}
