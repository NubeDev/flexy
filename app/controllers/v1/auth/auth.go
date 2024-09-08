package authController

import (
	"log"
	"net/http"

	model "github.com/NubeDev/flexy/app/models"
	userService "github.com/NubeDev/flexy/app/services/v1/user"
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils"
	"github.com/NubeDev/flexy/utils/code"

	"github.com/gin-gonic/gin"
)

var ExpireTimeFormat = "2006/01/02 15:04:05"

func UserLogin(c *gin.Context) {
	appG := common.Gin{C: c}

	// Bind Payload to Struct
	var userLogin userService.AuthStruct
	if err := c.ShouldBindJSON(&userLogin); err != nil {
		appG.Response(http.StatusBadRequest, code.InvalidParams, err.Error(), nil)
		return
	}

	// Validate Bound Struct Parameters
	err, parameterErrorStr := common.CheckBindStructParameter(userLogin, c)
	if utils.HandleError(c, http.StatusBadRequest, code.InvalidParams, parameterErrorStr, err) {
		return
	}

	data := make(map[string]interface{})
	RCode := code.InvalidParams
	isExist, userId, roleKey, isAdmin, status := model.CheckAuth(userLogin.Username, userLogin.Password)
	if !isExist {
		RCode = code.ErrorUserPasswordInvalid
		appG.Response(http.StatusOK, RCode, code.GetMsg(RCode), data)
		return
	}

	if status == 0 {
		appG.Response(http.StatusOK, code.ErrorAuth, "The user has been disabled", data)
		return
	}

	username := userLogin.Username
	claims := utils.Claims{
		UserId:   userId,
		Username: username,
		RoleKey:  roleKey,
		IsAdmin:  isAdmin,
	}

	accessToken, expireTime, err := utils.GenerateToken(claims)
	if utils.HandleError(c, http.StatusOK, code.AccessTokenFailure, code.GetMsg(code.AccessTokenFailure), err) {
		log.Println("Error generating access token: ", err)
		return
	}

	// Implement and assign refresh token
	refreshToken, _, refreshErr := utils.GenerateRefreshToken(claims)
	if utils.HandleError(c, http.StatusOK, code.RefreshAccessTokenFailure, code.GetMsg(code.RefreshAccessTokenFailure), refreshErr) {
		log.Println("Error generating refresh token: ", refreshErr)
		return
	}

	// Set the logged-in user information
	userService.SetLoggedUserInfo(userId, refreshToken)

	// Prepare the response data
	data["accessToken"] = accessToken
	data["token"] = accessToken
	data["refreshToken"] = refreshToken
	data["username"] = username
	data["nickname"] = username
	data["roles"] = [1]string{roleKey}
	data["expires"] = expireTime.Format(ExpireTimeFormat)

	RCode = code.SUCCESS
	appG.Response(http.StatusOK, RCode, "User login successful", data)
}

func RefreshAccessToken(c *gin.Context) {
	appG := common.Gin{C: c}
	var refreshAccessTokenStruct userService.RefreshAccessTokenStruct
	if err := c.ShouldBindJSON(&refreshAccessTokenStruct); err != nil {
		appG.Response(http.StatusBadRequest, code.InvalidParams, err.Error(), nil)
		return
	}

	data, err := userService.RefreshAccessToken(refreshAccessTokenStruct.RefreshToken)
	if utils.HandleError(c, http.StatusOK, code.ErrorAuthToken, "Failed to refresh access_token", err) {
		log.Println("Error token: ", err)
		return
	}

	appG.Response(http.StatusOK, code.SUCCESS, "Access token refreshed successfully!", data)
}

func UserLogout(c *gin.Context) {
	appG := common.Gin{C: c}
	claims, _ := c.Get("claims")
	user := claims.(*utils.Claims)

	userService.JoinBlockList(user.UserId, c.GetHeader("Authorization")[7:])
	appG.Response(http.StatusOK, code.SUCCESS, "ok", nil)
}

func ChangePassword(c *gin.Context) {
	appG := common.Gin{C: c}

	var userChangePassword userService.ChangePasswordStruct
	err := c.ShouldBindJSON(&userChangePassword)

	if utils.HandleError(c, http.StatusBadRequest, http.StatusBadRequest, "Parameter binding failed", err) {
		return
	}

	err, parameterErrorStr := common.CheckBindStructParameter(userChangePassword, c)
	if err != nil {
		appG.Response(http.StatusBadRequest, code.InvalidParams, parameterErrorStr, nil)
		return
	}

	claims, _ := c.Get("claims")
	user := claims.(*utils.Claims)

	isExist, userId, _, _, status := model.CheckAuth(user.Username, userChangePassword.OldPassword)
	if !isExist {
		RCode := code.ErrorUserOldPasswordInvalid
		appG.Response(http.StatusOK, RCode, code.GetMsg(RCode), nil)
		return
	}

	if status == 0 {
		appG.Response(http.StatusOK, code.ErrorAuth, "The user has been disabled", nil)
		return
	}

	passwordChangeSuccessful := userService.ChangeUserPassword(userId, userChangePassword.NewPassword)
	if !passwordChangeSuccessful {
		appG.Response(http.StatusOK, code.UnknownError, code.GetMsg(code.UnknownError), nil)
		return
	}

	userService.JoinBlockList(user.UserId, c.GetHeader("Authorization")[7:])
	appG.Response(http.StatusOK, code.SUCCESS, code.GetMsg(code.SUCCESS), nil)
}

func GetLoggedInUser(c *gin.Context) {
	appG := common.Gin{C: c}

	claims, _ := c.Get("claims")
	user := claims.(*utils.Claims)

	data := make(map[string]interface{}, 0)
	data["user_id"] = user.UserId
	data["user_name"] = user.Username
	data["roles"] = [...]string{user.RoleKey}
	data["permissions"] = [...]string{""}
	if user.IsAdmin {
		data["permissions"] = [...]string{"*:*:*"}
	}
	appG.Response(http.StatusOK, code.SUCCESS, "Current logged-in user information", data)
}
