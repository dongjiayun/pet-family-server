package router

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go-pet-family/controllers"
	"go-pet-family/models"
)

func genDoc(router *gin.Engine) {
	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
}

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

	r.Use(checkUserExtendInfoMiddleware)

	r.GET(":cid", controllers.GetUser)

	r.PUT("", controllers.UpdateUser)

	r.DELETE(":cid", controllers.DeleteUser)

	r.DELETE("/delete/:cid", controllers.HardDeleteUser)

	r.POST("/follow/:cid", controllers.FollowUser)

	r.DELETE("/follow/:cid", controllers.UnFollowUser)

	r.POST("", controllers.CreateUser)

	r.POST("get", controllers.GetUsers)

	r.POST("/myLikeArticles", controllers.MyLikeArticles)

	r.POST("/myCollectArticles", controllers.MyCollects)
}

func getArticleApi(router *gin.Engine) {
	r := router.Group("/article")

	r.Use(checkUserExtendInfoMiddleware)

	r.POST("get", controllers.GetArticles)

	r.GET(":articleId", controllers.GetArticle)

	r.Use(checkTokenMiddleware)

	r.POST("", controllers.CreateArticle)

	r.PUT("", controllers.UpdateArticle)

	r.DELETE(":articleId", controllers.DeleteArticle)

	r.POST("/like/:articleId", controllers.LikeArticle)

	r.DELETE("/like/:articleId", controllers.CancelLikeArticle)

	r.POST("/collect/:articleId", controllers.CollectArticle)

	r.DELETE("/collect/:articleId", controllers.CancelCollectArticle)

	r.GET("/checkLikeAndCollect/:articleId", controllers.CheckLikeAndCollect)

	r.POST("/private/:articleId", controllers.SetArticlePrivate)
}

func getTagApi(router *gin.Engine) {
	r := router.Group("/tag")

	r.Use(checkTokenMiddleware)

	r.Use(checkUserExtendInfoMiddleware)

	r.POST("get", controllers.GetTags)

	r.GET(":tagId", controllers.GetTag)

	r.POST("", controllers.CreateTag)

	r.PUT("", controllers.UpdateTag)

	r.DELETE(":tagId", controllers.DeleteTag)
}

func getCommentApi(router *gin.Engine) {
	r := router.Group("/comment")

	r.Use(checkTokenMiddleware)

	r.POST("get", controllers.GetComments)

	r.GET(":commentId", controllers.GetComment)

	r.POST("", controllers.CreateComment)

	r.PUT("", controllers.UpdateComment)

	r.DELETE(":commentId", controllers.DeleteComment)
}

func getNoticeApi(router *gin.Engine) {
	r := router.Group("/notice")

	r.Use(checkTokenMiddleware)

	r.POST("get", controllers.GetNotices)

	r.PUT("/:noticeId", controllers.ReadNotice)

	r.PUT("/readAll", controllers.ReadAllNotices)
}

func getCommonApi(router *gin.Engine) {
	r := router.Group("/common")

	r.GET("getAllArea", controllers.GetAllArea)

	r.POST("obsToken", controllers.GetObsToken)

	r.Use(checkTokenMiddleware)
}

func GetRouter() *gin.Engine {
	router := gin.Default()
	genDoc(router)
	getAuthApi(router)
	getUserApi(router)
	getArticleApi(router)
	getTagApi(router)
	getCommentApi(router)
	getNoticeApi(router)
	getCommonApi(router)
	return router
}
