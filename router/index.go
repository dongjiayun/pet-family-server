package router

import (
	"github.com/gin-gonic/gin"
	"go-pet-family/controllers"
)

func getAuthApi(router *gin.Engine) {
	r := router.Group("/auth")

	r.POST("/login", controllers.SignIn)
}

func getUserApi(router *gin.Engine) {
	r := router.Group("/user")

	r.POST("", controllers.CreateUser)

	r.GET("", controllers.GetUsers)
}

func GetRouter() *gin.Engine {
	router := gin.Default()
	getAuthApi(router)
	getUserApi(router)
	return router
}
