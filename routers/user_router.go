package routers

import (
	authController "github.com/NubeDev/flexy/app/controllers/v1/auth"
	userController "github.com/NubeDev/flexy/app/controllers/v1/user"
	"github.com/NubeDev/flexy/app/middleware"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(Router *gin.RouterGroup) {
	Router.POST("/login", authController.UserLogin)                  // Login
	Router.POST("/refresh_token", authController.RefreshAccessToken) // Refresh access_token
	Router.POST("/user", userController.CreateUser)                  // Create user

	user := Router.Group("/user").Use(
		middleware.TranslationHandler(),
		middleware.JWTHandler(),
		middleware.CasbinHandler())
	{
		//user.POST("", userController.CreateUser)                    // Create user
		user.GET("", userController.GetUsers)                       // User list
		user.PUT("/logout", authController.UserLogout)              // Logout
		user.PUT("/change_password", authController.ChangePassword) // Change password
		user.GET("/logged_in", authController.GetLoggedInUser)      // Current logged-in user information
	}
}
