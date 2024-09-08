package middleware

import (
	"fmt"
	"net/http"

	"github.com/NubeDev/flexy/utils"
	"github.com/NubeDev/flexy/utils/casbin"

	"github.com/gin-gonic/gin"
)

func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.Get("claims")
		user := claims.(*utils.Claims)

		username := user.Username
		roleKey := user.RoleKey
		obj := c.Request.URL.EscapedPath()
		act := c.Request.Method

		// Check route permissions
		check, _ := casbin.CasbinEnforcer.Enforce(roleKey, obj, act)
		if !check {
			// log.Println("Permission check failed")
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  fmt.Sprintf("[%s] does not have [%s] permissions for the [%s] route [%s]", username, roleKey, obj, act),
				"data": gin.H{},
			})
			c.Abort()
			return
		}

		c.Next()
		//log.Println(user, obj, act)
	}
}
