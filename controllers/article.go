package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
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
	var articles []models.Article
	db := models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&articles)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success", Data: articles})
			return
		}
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	for _, article := range articles {
		syncArticleInfo(&article)
	}

	var count int64
	models.DB.Model(&articles).Count(&count)

	list := models.GetListData[models.Article](articles, pageNo, pageSize, count)

	c.JSON(200, models.Result{0, "success", list})
}

func GetArticle(c *gin.Context) {
	type ArticleReq struct {
		Sync bool `json:"sync"`
	}
	var articleReq ArticleReq
	err := c.ShouldBindJSON(&articleReq)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
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
	syncArticleInfo(&article)
	c.JSON(200, models.Result{0, "success", article})
}

func syncArticleInfo(article *models.Article) {
	cid := article.Author.Cid
	var author models.SafeUser
	models.DB.Where("cid = ?", cid).First(&author)
	article.Author = author
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

	db := models.DB.Create(&article)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", article.ArticleId})
}

func UpdateArticle(c *gin.Context) {
	articleId := c.Param("articleId")
	var article models.Article
	err := c.ShouldBindJSON(&article)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	db := models.DB.Model(&article).
		Omit("Author", "Covers", "Tags", "Location", "Likes", "Collects", "Comments").
		Where("article_id = ?", articleId).
		Updates(&article)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", article.ArticleId})
}
