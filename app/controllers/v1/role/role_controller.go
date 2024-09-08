package roleController

import (
	"net/http"

	roleService "github.com/NubeDev/flexy/app/service/v1/role"
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/com"
	"github.com/gin-gonic/gin"
)

func DeleteRole(c *gin.Context) {
	appG := common.Gin{C: c}
	roleId, err := com.StrTo(c.Param("role_id")).Uint()
	if utils.HandleError(c, http.StatusBadRequest, http.StatusBadRequest, "Parameter binding failed", err) {
		return
	}

	deleteSuccessful := roleService.DeleteRole(roleId)
	if !deleteSuccessful {
		appG.Response(http.StatusOK, code.UnknownError, code.GetMsg(code.UnknownError), nil)
		return
	}

	appG.Response(http.StatusOK, code.SUCCESS, "ok", nil)
}

func CreateRole(c *gin.Context) {
	appG := common.Gin{C: c}

	var createRole roleService.CreateRoleStruct
	err := c.ShouldBindJSON(&createRole)

	if utils.HandleError(c, http.StatusBadRequest, http.StatusBadRequest, "Parameter binding failed", err) {
		return
	}

	err, parameterErrorStr := common.CheckBindStructParameter(createRole, c)
	if utils.HandleError(c, http.StatusBadRequest, code.InvalidParams, parameterErrorStr, err) {
		return
	}

	err = roleService.CreateRole(createRole)
	if utils.HandleError(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to add new role!", err) {
		return
	}

	appG.Response(http.StatusOK, code.SUCCESS, "ok", createRole)
}

func UpdateRole(c *gin.Context) {
	appG := common.Gin{C: c}
	roleId := com.StrTo(c.Param("role_id")).MustInt()

	var updateRole roleService.UpdateRoleStruct
	err := c.ShouldBindJSON(&updateRole)

	if utils.HandleError(c, http.StatusBadRequest, http.StatusBadRequest, "Parameter binding failed", err) {
		return
	}

	changeSuccessful := roleService.UpdateRole(roleId, updateRole)
	if !changeSuccessful {
		appG.Response(http.StatusOK, code.UnknownError, code.GetMsg(code.UnknownError), nil)
		return
	}

	appG.Response(http.StatusOK, code.SUCCESS, "ok", updateRole)
}

func GetRoles(c *gin.Context) {
	appG := common.Gin{C: c}
	err, errStr, p, n := utils.GetPage(c)
	if utils.HandleError(c, http.StatusBadRequest, code.InvalidParams, errStr, err) {
		return
	}

	roleServiceObj := roleService.RoleStruct{
		PageNum:  p,
		PageSize: n,
	}

	total, err := roleServiceObj.Count()
	if utils.HandleError(c, http.StatusInternalServerError, code.ERROR, "Failed to get total count", err) {
		return
	}

	roleArr, err := roleServiceObj.GetAll()
	if utils.HandleError(c, http.StatusInternalServerError, code.ERROR, "Server error", err) {
		return
	}

	data := utils.PageResult{
		List:     roleArr,
		Total:    total,
		PageSize: n,
	}
	appG.Response(http.StatusOK, code.SUCCESS, "ok", data)
}
