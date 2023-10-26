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

	r.Use(checkTokenMiddleware).POST("/signOut", controllers.SignOut)

	r.POST("/refreshToken", controllers.RefreshToken)
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

func getArticleApi(router *gin.Engine) {
	r := router.Group("/article")

	r.Use(checkTokenMiddleware)

	r.GET("", controllers.GetArticles)

	r.GET(":aid", controllers.GetArticle)

	r.POST("", controllers.CreateArticle)
}

func getTagApi(router *gin.Engine) {
	r := router.Group("/tag")

	r.Use(checkTokenMiddleware)

	r.GET("", controllers.GetTags)

	r.GET(":tid", controllers.GetTag)

	r.POST("", controllers.CreateTag)

	r.PUT(":tid", controllers.UpdateTag)
}

func GetRouter() *gin.Engine {
	router := gin.Default()
	getAuthApi(router)
	getUserApi(router)
	getArticleApi(router)
	getTagApi(router)
	return router
}
