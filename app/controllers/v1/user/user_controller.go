package userController

import (
	"net/http"

	userService "github.com/NubeDev/flexy/app/service/v1/user"
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	appG := common.Gin{C: c}

	var newUser userService.AddUserStruct
	err := c.ShouldBindJSON(&newUser)
	if utils.HandleError(c, http.StatusBadRequest, code.InvalidParams, "Parameter binding failed", err) {
		return
	}

	err, parameterErrorStr := common.CheckBindStructParameter(newUser, c)
	if utils.HandleError(c, http.StatusBadRequest, code.InvalidParams, parameterErrorStr, err) {
		return
	}

	err = userService.CreateUser(newUser)
	if utils.HandleError(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to add new user！", err) {
		return
	}

	appG.Response(http.StatusOK, code.SUCCESS, "User added successfully", nil)
}

func GetUsers(c *gin.Context) {
	appG := common.Gin{C: c}

	_, errStr, p, n := utils.GetPage(c)
	if errStr != "" {
		appG.Response(http.StatusBadRequest, code.InvalidParams, errStr, nil)
		return
	}

	var userServiceObj userService.UserStruct
	userServiceObj.PageNum = p
	userServiceObj.PageSize = n

	err := c.ShouldBindQuery(&userServiceObj)
	if utils.HandleError(c, http.StatusBadRequest, http.StatusBadRequest, "Parameter binding failed", err) {
		return
	}

	total, err := userServiceObj.Count()
	if err != nil {
		appG.Response(http.StatusInternalServerError, code.ERROR, "Failed to get total count: "+err.Error(), nil)
		return
	}

	userArr, err := userServiceObj.GetAll()
	if err != nil {
		appG.Response(http.StatusInternalServerError, code.ERROR, "Server error: "+err.Error(), nil)
		return
	}
	data := utils.PageResult{
		List:        userArr,
		Total:       total,
		CurrentPage: p,
		PageSize:    n,
	}
	appG.Response(http.StatusOK, code.SUCCESS, "ok", data)
}
