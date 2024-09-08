package routers

import (
	authController "github.com/NubeDev/flexy/app/controllers/v1/auth"
	userController "github.com/NubeDev/flexy/app/controllers/v1/user"
	"github.com/NubeDev/flexy/app/middleware"
	"github.com/gin-gonic/gin"
)

func InitUserRouter(router *gin.RouterGroup) {
	router.POST("/login", authController.UserLogin)                  // Login
	router.POST("/refresh_token", authController.RefreshAccessToken) // Refresh access_token
	router.POST("/users", userController.CreateUser)                 // Create user

	endPoint := router.Group("users").Use(middleware.TranslationHandler())
	if useAuth {
		endPoint.Use(
			middleware.JWTHandler(),
			middleware.CasbinHandler(),
		)
	}
	{
		endPoint.GET("", userController.GetUsers)                       // User list
		endPoint.PUT("/logout", authController.UserLogout)              // Logout
		endPoint.PUT("/change_password", authController.ChangePassword) // Change password
		endPoint.GET("/logged_in", authController.GetLoggedInUser)      // Current logged-in user information
	}
}
