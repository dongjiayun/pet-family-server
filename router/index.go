package router

import (
	"github.com/gin-gonic/gin"
	"go-pet-family/controllers"
	"go-pet-family/models"
)

func checkTokenMiddleware(c *gin.Context) {
	checkToken, _ := controllers.CheckToken(c)
	if checkToken == nil {
		c.JSON(403, models.Result{Code: 10001, Message: "token is invalid"})
		c.Abort()
		return
	}
	c.Set("cid", checkToken.Cid)
}

func getAuthApi(router *gin.Engine) {
	r := router.Group("/auth")

	r.POST("/signIn", controllers.SignIn)

	r.POST("/sendEmailOtp", controllers.SendEmailOtp)

	r.POST("/refreshToken", controllers.RefreshToken)

	r.Use(checkTokenMiddleware).POST("/signOut", controllers.SignOut)
}

func getUserApi(router *gin.Engine) {
	r := router.Group("/user")

	r.Use(checkTokenMiddleware)

	r.POST("", controllers.CreateUser)

	r.GET("", controllers.GetUsers)

	r.GET(":cid", controllers.GetUser)

	r.PUT(":cid", controllers.UpdateUser)

	r.DELETE(":cid", controllers.DeleteUser)

	r.DELETE("/delete/:cid", controllers.HardDeleteUser)
}

func GetRouter() *gin.Engine {
	router := gin.Default()
	getAuthApi(router)
	getUserApi(router)
	return router
}
