package middleware

import (
	"errors"
	"net/http"

	userService "github.com/NubeDev/flexy/app/services/v1/user"
	"github.com/NubeDev/flexy/utils"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var rCode int
		var data interface{}

		rCode = code.SUCCESS

		var token = c.Query("token") // add as a parm or add as a Header
		var authorization = c.GetHeader("token")
		if authorization != "" {
			token = authorization
		}

		var claims *utils.Claims
		var err error
		if token == "" {
			rCode = code.TokenInvalid
		} else {
			claims, err = utils.ParseToken(token)

			if err != nil {
				rCode = code.ErrorAuthCheckTokenFail
				if errors.Is(err, jwt.ErrTokenMalformed) {
					// log.Println("That's not even a token")
				} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
					// log.Println("Invalid signature")
				} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
					rCode = code.ErrorAuthCheckTokenTimeout
					// log.Println("Timing is everything")
				}
			}
		}

		jwtCount, _ := userService.InBlockList(token)
		if jwtCount >= 1 {
			rCode = code.AuthTokenInBlockList
		}

		if rCode != code.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": rCode,
				"msg":  code.GetMsg(rCode),
				"data": data,
			})

			c.Abort()
			return
		}
		// Store login user information
		c.Set("claims", claims)
		c.Next()
	}
}
