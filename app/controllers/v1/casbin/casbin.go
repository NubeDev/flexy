package casbinController

import (
	casbinService "github.com/NubeDev/flexy/app/services/v1/casbin"
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils"
	"github.com/NubeDev/flexy/utils/casbin"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/com"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateCasbin(c *gin.Context) {
	appG := common.Gin{C: c}

	var newCasbin casbinService.AddCasbinStruct
	err := c.ShouldBindJSON(&newCasbin)
	if utils.HandleError(c, http.StatusBadRequest, code.InvalidParams, "Parameter binding failed", err) {
		return
	}

	err, parameterErrorStr := common.CheckBindStructParameter(newCasbin, c)
	if utils.HandleError(c, http.StatusBadRequest, code.InvalidParams, parameterErrorStr, err) {
		return
	}

	err = casbinService.CreateCasbin(newCasbin)
	if utils.HandleError(c, http.StatusInternalServerError, http.StatusInternalServerError, "Path addition failed!", err) {
		return
	}

	// Regenerate the permission list
	casbin.SetupCasbin()
	appG.Response(http.StatusOK, code.SUCCESS, "Path added successfully", nil)
}

func UpdateCasbin(c *gin.Context) {
	appG := common.Gin{C: c}
	id := com.StrTo(c.Param("id")).MustInt()

	var update casbinService.AddCasbinStruct
	err := c.ShouldBindJSON(&update)

	if utils.HandleError(c, http.StatusBadRequest, http.StatusBadRequest, "Parameter binding failed", err) {
		return
	}

	changeSuccessful := casbinService.UpdateCasbin(id, update)
	if !changeSuccessful {
		appG.Response(http.StatusOK, code.UnknownError, code.GetMsg(code.UnknownError), nil)
		return
	}

	// Regenerate the permission list
	casbin.SetupCasbin()

	appG.Response(http.StatusOK, code.SUCCESS, "ok", update)
}

func DeleteCasbin(c *gin.Context) {
	appG := common.Gin{C: c}
	//id := com.StrTo(c.Param("id")).MustInt()

	var update casbinService.AddCasbinStruct
	err := c.ShouldBindJSON(&update)

	if utils.HandleError(c, http.StatusBadRequest, http.StatusBadRequest, "Parameter binding failed", err) {
		return
	}

	// Regenerate the permission list
	casbin.SetupCasbin().RemovePolicy(update.V0, update.V1, update.V2)

	appG.Response(http.StatusOK, code.SUCCESS, "ok", update)
}

func GetCasbinList(c *gin.Context) {
	appG := common.Gin{C: c}
	groupBy := c.DefaultQuery("group_by", "")
	err, errStr, p, n := utils.GetPage(c)
	if utils.HandleError(c, http.StatusBadRequest, code.InvalidParams, errStr, err) {
		return
	}

	casbinServiceObj := casbinService.CasbinStruct{
		PageNum:  p,
		PageSize: n,
		V0:       c.DefaultQuery("role", ""),
		V1:       c.DefaultQuery("path", ""),
		V2:       c.DefaultQuery("method", ""),
	}

	total, err := casbinServiceObj.Count()
	if utils.HandleError(c, http.StatusInternalServerError, code.ERROR, "Failed to get page count", err) {
		return
	}

	arr, err := casbinServiceObj.GetAll()
	if utils.HandleError(c, http.StatusInternalServerError, code.ERROR, "Server error", err) {
		return
	}

	data := utils.PageResult{
		List:     arr,
		Total:    total,
		PageSize: n,
	}

	// Filter group by
	if groupBy == "v0" {
		groupMap := make(map[string][]interface{})
		for k, v := range arr {
			if arr[k].V0 == v.V0 {
				groupMap[arr[k].V0] = append(groupMap[arr[k].V0], v)
			}
		}
		data.List = groupMap
	}

	appG.Response(http.StatusOK, code.SUCCESS, "ok", data)
}
