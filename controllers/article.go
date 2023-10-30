package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
	"go-pet-family/utils"
	"time"
)

type ArticlesReq struct {
	models.Pagination
	Sync bool `json:"sync"`
}

func GetArticles(c *gin.Context) {
	articlesReq := ArticlesReq{
		Pagination: models.Pagination{
			PageSize: 20,
			PageNo:   1,
		},
	}
	err := c.ShouldBindJSON(&articlesReq)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	pageNo := articlesReq.PageNo
	pageSize := articlesReq.PageSize
	var articles models.Articles
	db := models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&articles)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	if articlesReq.Sync {
		ch := make(chan error)
		go syncArticleInfos(&articles, ch)
		<-ch
	}

	var count int64
	models.DB.Model(&articles).Count(&count)

	list := models.GetListData[models.Article](articles, pageNo, pageSize, count)

	c.JSON(200, models.Result{0, "success", list})
}

func GetArticle(c *gin.Context) {
	aid := c.Param("articleId")
	var article models.Article
	db := models.DB.Where("article_id = ?", aid).First(&article)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	ch := make(chan error)
	go syncArticleInfo(&article, ch)
	<-ch
	c.JSON(200, models.Result{0, "success", article})
}

func syncArticleInfo(article *models.Article, ch chan error) {
	cid := article.Author.Cid
	var author models.User
	models.DB.Where("cid = ?", cid).First(&author)
	article.Author = models.GetSafeUser(author)
	var tagIds []string
	for _, tag := range article.Tags {
		tagIds = append(tagIds, tag.TagId)
	}
	var tags []models.Tag
	models.DB.Where("tag_id in (?)", tagIds).Find(&tags)
	article.Tags = tags
	models.DB.Model(&article).Where("article_id = ?", article.ArticleId).Updates(&article)
	<-ch
}

func syncArticleInfos(articles *models.Articles, ch chan error) {
	var cids []string
	var users models.Users
	for _, article := range *articles {
		cids = append(cids, article.Author.Cid)
		var tagIds []string
		for _, tag := range article.Tags {
			tagIds = append(tagIds, tag.TagId)
		}
		var tags []models.Tag
		article.Tags = tags
		models.DB.Where("tag_id in (?)", tagIds).Find(&tags)
	}
	models.DB.Where("cid in (?)", cids).Find(&users)
	for index, article := range *articles {
		user := models.GetSafeUser(users[index])
		article.Author = user
	}
	models.DB.Where("article_id in (?)", cids).Updates(&articles)
	ch <- nil
}

func CreateArticle(c *gin.Context) {
	cid, _ := c.Get("cid")
	var article models.Article
	err := c.ShouldBindJSON(&article)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	if cid != nil {
		article.Author.Cid = cid.(string)
	}

	uuid := uuid.New()
	uuidStr := uuid.String()
	article.ArticleId = "Article-" + uuidStr

	article.IsAudit = true
	article.AuditBy = "C-ADMIN"
	article.AuditAt = time.Now()

	db := models.DB.Create(&article)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", article.ArticleId})
}

func UpdateArticle(c *gin.Context) {
	var requestBody models.Article
	err := c.ShouldBindJSON(&requestBody)
	articleId := requestBody.ArticleId
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	article := models.Article{}
	db := models.DB.Where("article_id = ?", articleId).Where("deleted_at IS NULL").First(&article)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	update := models.Article{
		Title:    requestBody.Title,
		Content:  requestBody.Content,
		Covers:   requestBody.Covers,
		Tags:     requestBody.Tags,
		Location: requestBody.Location,
	}

	update.IsAudit = true
	update.AuditBy = "C-ADMIN"
	update.AuditAt = time.Now()

	cid, _ := c.Get("cid")
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	update.UpdateBy = user.Cid

	updateDb := models.DB.Model(&update).
		Where("article_id = ?", articleId).
		Updates(&update)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	c.JSON(200, models.Result{0, "success", article.ArticleId})
}

func LikeArticle(c *gin.Context) {
	articleId := c.Param("articleId")
	var article models.Article
	db := models.DB.Where("article_id = ?", articleId).First(&article)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	cid, _ := c.Get("cid")
	var userExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid.(string)).First(&userExtendInfo)
	hasLike := utils.ArrayIncludes[models.Article](userExtendInfo.LikeArticles, articleId, func(item models.Article) any {
		return item.ArticleId
	})
	if hasLike {
		c.JSON(200, models.Result{Code: 10001, Message: "您已点赞"})
		return
	}
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	like := models.GetSafeUser(user)
	article.Likes = append(article.Likes, like)
	updateDb := models.DB.Model(&article).Where("article_id = ?", articleId).Update("likes", article.Likes)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	userExtendInfo.LikeArticles = append(userExtendInfo.LikeArticles, article)
	updateDb = models.DB.Model(&userExtendInfo).Where("cid = ?", cid.(string)).Update("like_articles", userExtendInfo.LikeArticles)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func CancelLikeArticle(c *gin.Context) {
	articleId := c.Param("articleId")
	var article models.Article
	db := models.DB.Where("article_id = ?", articleId).First(&article)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	cid, _ := c.Get("cid")
	var userExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid.(string)).First(&userExtendInfo)
	hasLike := utils.ArrayIncludes[models.Article](userExtendInfo.LikeArticles, articleId, func(item models.Article) any {
		return item.ArticleId
	})
	if hasLike == false {
		c.JSON(200, models.Result{Code: 10001, Message: "您已取消"})
		return
	}
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	like := models.GetSafeUser(user)
	likes := utils.ArrayFilter[models.SafeUser](article.Likes, func(item models.SafeUser) bool {
		return item.Cid != like.Cid
	})
	article.Likes = likes
	updateModel := models.DB.Model(&article).Where("article_id = ?", articleId).Update("likes", article.Likes)
	if updateModel.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	userExtendInfo.LikeArticles = utils.ArrayFilter[models.Article](userExtendInfo.LikeArticles, func(item models.Article) bool {
		return item.ArticleId != articleId
	})
	updateModel = models.DB.Model(&userExtendInfo).Where("cid = ?", cid.(string)).Update("like_articles", userExtendInfo.LikeArticles)
	if updateModel.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func CollectArticle(c *gin.Context) {
	articleId := c.Param("articleId")
	var article models.Article
	db := models.DB.Where("article_id = ?", articleId).First(&article)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	cid, _ := c.Get("cid")
	var userExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid.(string)).First(&userExtendInfo)
	hasCollect := utils.ArrayIncludes[models.Article](userExtendInfo.Collects, articleId, func(item models.Article) any {
		return item.ArticleId
	})
	if hasCollect {
		c.JSON(200, models.Result{Code: 10001, Message: "您已收藏"})
		return
	}
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	collect := models.GetSafeUser(user)
	article.Collects = append(article.Collects, collect)
	updateDb := models.DB.Model(&article).Where("article_id = ?", articleId).Update("collects", article.Collects)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	userExtendInfo.Collects = append(userExtendInfo.Collects, article)
	updateDb = models.DB.Model(&userExtendInfo).Where("cid = ?", cid.(string)).Update("collects", userExtendInfo.Collects)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func CancelCollectArticle(c *gin.Context) {
	articleId := c.Param("articleId")
	var article models.Article
	db := models.DB.Where("article_id = ?", articleId).First(&article)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	cid, _ := c.Get("cid")
	var userExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid.(string)).First(&userExtendInfo)
	hasCollect := utils.ArrayIncludes[models.Article](userExtendInfo.Collects, articleId, func(item models.Article) any {
		return item.ArticleId
	})
	if hasCollect == false {
		c.JSON(200, models.Result{Code: 10001, Message: "您已取消"})
		return
	}
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	collect := models.GetSafeUser(user)
	article.Collects = utils.ArrayFilter[models.SafeUser](article.Collects, func(item models.SafeUser) bool {
		return item.Cid != collect.Cid
	})
	updateDb := models.DB.Model(&article).Where("article_id = ?", articleId).Update("collects", article.Collects)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	userExtendInfo.Collects = utils.ArrayFilter[models.Article](userExtendInfo.Collects, func(item models.Article) bool {
		return item.ArticleId != articleId
	})
	updateDb = models.DB.Model(&userExtendInfo).Where("cid = ?", cid.(string)).Update("collects", userExtendInfo.Collects)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}
