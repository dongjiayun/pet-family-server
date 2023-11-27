package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/importcjj/sensitive"
	"go-pet-family/models"
	"go-pet-family/utils"
	"time"
)

type ArticlesReq struct {
	models.Pagination
	Sync bool   `json:"sync"`
	Cid  string `json:"cid"`
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
	var cid string
	var reqCid string
	if articlesReq.Cid != "" {
		reqCid = articlesReq.Cid
	}
	if reqCid == "" {
		tokenString := c.GetHeader("Authorization")
		if tokenString != "" {
			checkToken, _ := CheckToken(c)
			if checkToken == nil {
				c.JSON(403, models.Result{Code: 10001, Message: "token is invalid"})
				c.Abort()
				return
			}
			cid = checkToken.Cid
		}
	}
	if cid == "C000000000001" {
		db := models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).
			Where("deleted_at IS NULL").Order("id desc").Find(&articles)
		if db.Error != nil {
			if db.Error.Error() == "record not found" {
				c.JSON(200, models.Result{Code: 0, Message: "success"})
				return
			}
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
			return
		}
	} else if cid != "" {
		cidStr := cid
		db := models.DB.Limit(pageSize).Offset((pageNo-1)*pageSize).
			Where("deleted_at IS NULL").
			Where("is_private = ?", false)
		if reqCid != "" {
			db.Where("author_id = ? ", reqCid)
		}
		db.Or("author_id = ? and deleted_at IS NULL", cidStr).
			Order("id desc")
		db.Find(&articles)
		if db.Error != nil {
			if db.Error.Error() == "record not found" {
				c.JSON(200, models.Result{Code: 0, Message: "success"})
				return
			}
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
			return
		}
	} else {
		db := models.DB.Limit(pageSize).Offset((pageNo-1)*pageSize).
			Where("deleted_at IS NULL").
			Where("is_private = ?", false)
		if reqCid != "" {
			db.Where("author_id = ? ", reqCid)
		}
		db.Order("id desc")
		db.Find(&articles)
		if db.Error != nil {
			if db.Error.Error() == "record not found" {
				c.JSON(200, models.Result{Code: 0, Message: "success"})
				return
			}
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
			return
		}
	}

	utils.ArrayForeach[models.Article]((*[]models.Article)(&articles), func(article models.Article) models.Article {
		article.ColllectCount = len(article.CollectCids)
		article.CommentCount = len(article.CommentIds)
		article.LikesCount = len(article.LikeCids)
		return article
	})

	if articlesReq.Sync {
		ch := make(chan error)
		go syncArticleInfos(&articles, ch)
		<-ch
	}

	var count int64

	models.DB.Debug().Model(&articles).Where("deleted_at IS NULL").Where("is_private = ?", false).Count(&count)

	list := models.GetListData[models.Article](articles, pageNo, pageSize, count)

	c.JSON(200, models.Result{0, "success", list})
}

func GetArticle(c *gin.Context) {
	aid := c.Param("articleId")
	var article models.Article
	var cid string
	tokenString := c.GetHeader("Authorization")
	if tokenString != "" {
		checkToken, _ := CheckToken(c)
		if checkToken == nil {
			c.JSON(403, models.Result{Code: 10001, Message: "token is invalid"})
			c.Abort()
			return
		}
		cid = checkToken.Cid
	}
	if cid == "C000000000001" {
		db := models.DB.
			Where("article_id = ?", aid).
			Where("deleted_at IS NULL").
			First(&article)
		if db.Error != nil {
			if db.Error.Error() == "record not found" {
				c.JSON(200, models.Result{Code: 0, Message: "success"})
				return
			}
			// SQL执行失败，返回错误信息
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
			return
		}
	} else if cid != "" {
		cidStr := cid
		db := models.DB.
			Where("article_id = ?", aid).
			Where("deleted_at IS NULL").
			Where("is_private = ?", false).
			Or("author_id = ? and deleted_at IS NULL and article_id = ?", cidStr, aid).
			First(&article)
		if db.Error != nil {
			if db.Error.Error() == "record not found" {
				c.JSON(200, models.Result{Code: 0, Message: "success"})
				return
			}
			// SQL执行失败，返回错误信息
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
			return
		}
	} else {
		db := models.DB.
			Where("article_id = ?", aid).
			Where("deleted_at IS NULL").
			Where("is_private = ?", false).
			First(&article)
		if db.Error != nil {
			if db.Error.Error() == "record not found" {
				c.JSON(200, models.Result{Code: 0, Message: "success"})
				return
			}
			// SQL执行失败，返回错误信息
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
			return
		}
	}
	ch := make(chan error)
	go syncArticleInfo(&article, ch, c)
	<-ch

	article.ColllectCount = len(article.CollectCids)
	article.CommentCount = len(article.CommentIds)
	article.LikesCount = len(article.LikeCids)

	c.JSON(200, models.Result{0, "success", article})
}

func syncArticleInfo(article *models.Article, ch chan error, c *gin.Context) {
	cid := article.Author.Cid
	var author models.User
	db := models.DB.Where("cid = ?", cid).First(&author)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	article.Author = models.GetSafeUser(author)
	var tagIds []string
	for _, tag := range article.Tags {
		tagIds = append(tagIds, tag.TagId)
	}
	var tags []models.Tag
	tagsDb := models.DB.Where("tag_id in (?)", tagIds).Find(&tags)
	if tagsDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	article.Tags = tags
	updateDb := models.DB.Model(&article).Where("article_id = ?", article.ArticleId).Updates(&article)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	ch <- nil
}

func syncArticleInfos(articles *models.Articles, ch chan error) {
	cids := []string{}
	for _, article := range *articles {
		cids = append(cids, article.ArticleId)
		var tagIds []string
		for _, tag := range article.Tags {
			tagIds = append(tagIds, tag.TagId)
		}
		var tags []models.Tag
		article.Tags = tags
		models.DB.Where("tag_id in (?)", tagIds).Find(&tags)
		var user models.User
		models.DB.Where("cid = ?", article.Author.Cid).First(&user)
		article.Author = models.GetSafeUser(user)
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
		article.AuthorId = cid.(string)
	}

	uuid := uuid.New()
	uuidStr := uuid.String()
	article.ArticleId = "Article-" + uuidStr

	filter := sensitive.New()
	filter.LoadWordDict("config/sensitiveDict.txt")

	isTitleSensitive, _ := filter.Validate(article.Title)

	if isTitleSensitive == false {
		c.JSON(200, models.Result{Code: 10002, Message: "文章标题存在敏感词😅"})
		return
	}

	isArticleSensitive, _ := filter.Validate(article.Content)
	if isArticleSensitive == false {
		c.JSON(200, models.Result{Code: 10002, Message: "文章内容存在敏感词😅"})
		return
	}

	models.CommonCreate[models.Article](&article)

	db := models.DB.Create(&article)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	SendArticleMessageToAllFollows(&article, c)
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

	var ch1 = make(chan string)
	go CheckSelfOrAdmin(c, article.AuthorId, ch1)
	message := <-ch1

	if message != "success" {
		c.JSON(200, models.Result{Code: 10001, Message: message})
		return
	}

	filter := sensitive.New()
	filter.LoadWordDict("config/sensitiveDict.txt")

	isTitleSensitive, _ := filter.Validate(requestBody.Title)

	if isTitleSensitive == false {
		c.JSON(200, models.Result{Code: 10002, Message: "文章标题存在敏感词😅"})
		return
	}

	isArticleSensitive, _ := filter.Validate(requestBody.Content)
	if isArticleSensitive == false {
		c.JSON(200, models.Result{Code: 10002, Message: "文章内容存在敏感词😅"})
		return
	}

	update := models.Article{
		Title:    requestBody.Title,
		Content:  requestBody.Content,
		Covers:   requestBody.Covers,
		Tags:     requestBody.Tags,
		Location: requestBody.Location,
	}

	ch := make(chan string)
	go models.CommonUpdate[models.Article](&update, c, ch)
	<-ch

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
	hasLike := utils.ArrayIncludes[string](userExtendInfo.LikeArticleIds, articleId, func(id string) any {
		return id
	})
	if hasLike {
		c.JSON(200, models.Result{Code: 10001, Message: "您已点赞"})
		return
	}
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	like := models.GetSafeUser(user)
	article.LikeCids = append(article.LikeCids, like.Cid)
	updateDb := models.DB.Model(&article).Where("article_id = ?", articleId).Update("like_cids", article.LikeCids)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	userExtendInfo.LikeArticleIds = append(userExtendInfo.LikeArticleIds, article.ArticleId)
	updateDb = models.DB.Model(&userExtendInfo).Where("cid = ?", cid.(string)).Update("like_article_ids", userExtendInfo.LikeArticleIds)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	title := "您有新的点赞"
	content := user.Username + "点赞了您的文章" + article.Title
	noticeType := "likeArticle"
	noticeId := article.ArticleId
	SendMessage(title, content, noticeType, noticeId, &user, article.Author.Cid, c)

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
	hasLike := utils.ArrayIncludes[string](userExtendInfo.LikeArticleIds, articleId, func(id string) any {
		return id
	})
	if hasLike == false {
		c.JSON(200, models.Result{Code: 10001, Message: "您已取消"})
		return
	}
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	like := models.GetSafeUser(user)
	likes := utils.ArrayFilter[string](article.LikeCids, func(id string) bool {
		return id != like.Cid
	})
	article.LikeCids = likes
	updateModel := models.DB.Model(&article).Where("article_id = ?", articleId).Update("like_cids", article.LikeCids)
	if updateModel.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	userExtendInfo.LikeArticleIds = utils.ArrayFilter[string](userExtendInfo.LikeArticleIds, func(id string) bool {
		return id != articleId
	})
	updateModel = models.DB.Model(&userExtendInfo).Where("cid = ?", cid.(string)).Update("like_article_ids", userExtendInfo.LikeArticleIds)
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
	hasCollect := utils.ArrayIncludes[string](userExtendInfo.CollectIds, articleId, func(id string) any {
		return id
	})
	if hasCollect {
		c.JSON(200, models.Result{Code: 10001, Message: "您已收藏"})
		return
	}
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	collect := models.GetSafeUser(user)
	article.CollectCids = append(article.CollectCids, collect.Cid)
	updateDb := models.DB.Model(&article).Where("article_id = ?", articleId).Update("collect_cids", article.CollectCids)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	userExtendInfo.CollectIds = append(userExtendInfo.CollectIds, article.ArticleId)
	updateDb = models.DB.Model(&userExtendInfo).Where("cid = ?", cid.(string)).Update("collect_ids", userExtendInfo.CollectIds)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	title := "您有新的收藏"
	content := user.Username + "收藏了您的文章" + article.Title
	noticeType := "collectArticle"
	noticeId := article.ArticleId
	SendMessage(title, content, noticeType, noticeId, &user, article.Author.Cid, c)

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
	hasCollect := utils.ArrayIncludes[string](userExtendInfo.CollectIds, articleId, func(id string) any {
		return id
	})
	if hasCollect == false {
		c.JSON(200, models.Result{Code: 10001, Message: "您已取消"})
		return
	}
	var user models.User
	models.DB.Where("cid = ?", cid.(string)).First(&user)
	collect := models.GetSafeUser(user)
	article.CollectCids = utils.ArrayFilter[string](article.CollectCids, func(id string) bool {
		return id != collect.Cid
	})
	updateDb := models.DB.Model(&article).Where("article_id = ?", articleId).Update("collect_cids", article.CollectCids)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	userExtendInfo.CollectIds = utils.ArrayFilter[string](userExtendInfo.CollectIds, func(id string) bool {
		return id != articleId
	})
	updateDb = models.DB.Model(&userExtendInfo).Where("cid = ?", cid.(string)).Update("collect_ids", userExtendInfo.CollectIds)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func CheckLikeAndCollect(c *gin.Context) {
	cid, _ := c.Get("cid")
	articleId := c.Param("articleId")
	var userExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid.(string)).First(&userExtendInfo)
	isLike := utils.ArrayIncludes[string](userExtendInfo.LikeArticleIds, articleId, func(id string) any {
		return id
	})
	isCollect := utils.ArrayIncludes[string](userExtendInfo.CollectIds, articleId, func(id string) any {
		return id
	})
	type Result struct {
		IsLike    bool `json:"isLike"`
		IsCollect bool `json:"isCollect"`
	}
	c.JSON(200, models.Result{Code: 0, Message: "success", Data: Result{IsLike: isLike, IsCollect: isCollect}})
}

func SetArticlePrivate(c *gin.Context) {
	type req struct {
		IsPrivate bool `json:"isPrivate"`
	}
	articleId := c.Param("articleId")
	var article models.Article

	var reqData req
	c.ShouldBindJSON(&reqData)
	db := models.DB.Where("article_id = ?", articleId).First(&article)

	var ch = make(chan string)
	go CheckSelfOrAdmin(c, article.AuthorId, ch)
	message := <-ch
	if message != "success" {
		c.JSON(200, models.Result{Code: 10001, Message: message})
		return
	}

	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	article.IsPrivate = reqData.IsPrivate
	updateDb := models.DB.Model(&article).Where("article_id = ?", articleId).Update("is_private", article.IsPrivate)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func DeleteArticle(c *gin.Context) {
	articleId := c.Param("articleId")
	var article models.Article
	db := models.DB.Where("article_id = ?", articleId).First(&article)

	var ch = make(chan string)
	go CheckSelfOrAdmin(c, article.AuthorId, ch)
	message := <-ch
	if message != "success" {
		c.JSON(200, models.Result{Code: 10001, Message: message})
		return
	}

	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	models.DB.Model(&article).Where("article_id = ?", articleId).Update("deleted_at", time.Now())
	c.JSON(200, models.Result{0, "success", nil})
}
