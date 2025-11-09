package handler

import (
	"cruder/internal/controller"

	"github.com/gin-gonic/gin"
)

func New(router *gin.Engine, userController *controller.UserController) *gin.Engine {
	v1 := router.Group("/api/v1")
	{
		userGroup := v1.Group("/users")
		{
			userGroup.GET("/", userController.GetAllUsers)
			userGroup.GET("/username/:username", userController.GetUserByUsername)
			userGroup.GET("/id/:id", userController.GetUserByID)
			userGroup.POST("/", userController.CreateUser)
			userGroup.PATCH("/:uuid", userController.UpdateUser)
			userGroup.DELETE("/:uuid", userController.DeleteUser)
		}
	}
	return router
}
