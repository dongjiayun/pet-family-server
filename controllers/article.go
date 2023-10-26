package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
)

func GetArticles(c *gin.Context) {
	pagination := models.Pagination{
		PageSize: 20,
		PageNo:   1,
	}
	err := c.ShouldBindJSON(&pagination)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	pageNo := pagination.PageNo
	pageSize := pagination.PageSize
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

	var count int64
	models.DB.Model(&articles).Count(&count)

	for i := range articles {
		article := &articles[i]
		setArticleInfo(article, c)
	}

	list := models.GetListData[models.Article](articles, pageNo, pageSize, count)

	c.JSON(200, models.Result{0, "success", list})
}

func GetArticle(c *gin.Context) {
	aid := c.Param("aid")
	var article models.Article
	db := models.DB.Where("article_id = ?", aid).First(&article)
	if db.Error != nil {
		fmt.Println(db.Error)
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	setArticleInfo(&article, c)
	c.JSON(200, models.Result{0, "success", article})
}

func setArticleInfo(article *models.Article, c *gin.Context) {
	var user models.User
	db := models.DB.Where("cid = ?", article.AuthorId).Where("deleted_at IS NULL").First(&user)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	article.Author = models.GetSafeUser(user)
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
		article.AuthorId = cid.(string)
	}

	uuid := uuid.New()
	uuidStr := uuid.String()
	article.ArticleId = "Article-" + uuidStr

	db := models.DB.Omit("Author", "Covers", "Tags", "Location", "Likes", "Collects", "Comments").Create(&article)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", article})
}
