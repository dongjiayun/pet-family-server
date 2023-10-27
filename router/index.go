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

func checkUserExtendInfoMiddleware(c *gin.Context) {
	cid, _ := c.Get("cid")
	cidString, _ := cid.(string)
	chUE := make(chan string)
	go controllers.CheckAndCreateExtendUserInfo(c, cidString, chUE)
	<-chUE
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

	r.Use(checkUserExtendInfoMiddleware)

	r.GET("", controllers.GetArticles)

	r.GET(":articleId", controllers.GetArticle)

	r.POST("", controllers.CreateArticle)

	r.PUT(":articleId", controllers.UpdateArticle)

	r.POST("/like/:articleId", controllers.LikeArticle)

	r.DELETE("/like/:articleId", controllers.CancelLikeArticle)
}

func getTagApi(router *gin.Engine) {
	r := router.Group("/tag")

	r.Use(checkTokenMiddleware)

	r.Use(checkUserExtendInfoMiddleware)

	r.GET("", controllers.GetTags)

	r.GET(":tagId", controllers.GetTag)

	r.POST("", controllers.CreateTag)

	r.PUT(":tagId", controllers.UpdateTag)

	r.DELETE(":tagId", controllers.DeleteTag)
}

func GetRouter() *gin.Engine {
	router := gin.Default()
	getAuthApi(router)
	getUserApi(router)
	getArticleApi(router)
	getTagApi(router)
	return router
}
