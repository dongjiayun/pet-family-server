package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/importcjj/sensitive"
	"go-pet-family/models"
	"go-pet-family/utils"
	"strings"
	"time"
)

type GetCommentsReq struct {
	models.Pagination
	TargetId  string `json:"targetId"`
	AuthorCid string `json:"cid"`
	Sync      bool   `json:"sync"`
}

func GetComments(c *gin.Context) {
	commentsReq := GetCommentsReq{
		Pagination: models.Pagination{
			PageSize: 20,
			PageNo:   1,
		},
	}

	err := c.ShouldBindJSON(&commentsReq)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}

	pageNo := commentsReq.PageNo
	pageSize := commentsReq.PageSize
	targetId := commentsReq.TargetId
	authorCid := commentsReq.AuthorCid

	fmt.Println("commentsReq", commentsReq)

	var comments models.Comments

	db := models.DB

	if targetId != "" {
		db = db.Where("target_id = ?", targetId)
	}

	if authorCid != "" {
		db = db.Where("author_id = ?", authorCid)
	}

	db.Debug().Order("id desc").
		Where("deleted_at IS NULL").
		Limit(pageSize).Offset((pageNo - 1) * pageSize).
		Find(&comments)

	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	var count int64

	models.DB.Table("comments").Where("target_id = ?", targetId).Where("deleted_at IS NULL").Count(&count)

	list := models.GetListData[models.Comment](comments, pageNo, pageSize, count)

	if commentsReq.Sync {
		ch := make(chan error)
		go syncCommentInfos(&comments, ch)
		<-ch
	}

	c.JSON(200, models.Result{0, "success", list})
}

func GetComment(c *gin.Context) {
	commentId := c.Param("commentId")
	var comment models.Comment
	db := models.DB.Where("comment_id = ?", commentId).Where("deleted_at IS NULL").First(&comment)
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
	go syncCommentInfo(&comment, ch)
	<-ch
	c.JSON(200, models.Result{0, "success", comment})
}

func UpdateComment(c *gin.Context) {
	var requestBody models.Comment
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}

	filter := sensitive.New()
	filter.LoadWordDict("config/sensitiveDict.txt")

	isSensitive, _ := filter.Validate(requestBody.Content)

	if isSensitive == false {
		c.JSON(200, models.Result{Code: 10002, Message: "评论存在敏感词😅"})
		return
	}

	commentId := requestBody.CommentId
	var oldComment models.Comment
	db := models.DB.Where("comment_id = ?", commentId).Where("deleted_at IS NULL").First(&oldComment)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	update := models.Comment{
		Content:     requestBody.Content,
		Location:    requestBody.Location,
		Attachments: requestBody.Attachments,
	}

	ch := make(chan string)
	go models.CommonUpdate[models.Comment](&update, c, ch)
	<-ch

	db = models.DB.Model(&oldComment).Where("comment_id = ?", commentId).Updates(&update)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", commentId})
}

func CreateComment(c *gin.Context) {
	var comment models.Comment
	err := c.ShouldBindJSON(&comment)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	cid, _ := c.Get("cid")
	if cid != nil {
		comment.Author.Cid = cid.(string)
	}
	uuid := uuid.New()
	uuidStr := uuid.String()
	comment.CommentId = "Comment-" + uuidStr

	isArticle := strings.Contains(comment.TargetId, "Article")

	if isArticle == false {
		if comment.RootCommentId == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "请传入根评论id"})
			return
		}
	}

	filter := sensitive.New()
	filter.LoadWordDict("config/sensitiveDict.txt")

	isSensitive, _ := filter.Validate(comment.Content)

	if isSensitive == false {
		c.JSON(200, models.Result{Code: 10002, Message: "评论存在敏感词😅"})
		return
	}

	var user models.User
	models.DB.Where("cid = ?", cid).First(&user)

	models.CommonCreate[models.Comment](&comment)

	safeUser := models.GetSafeUser(user)

	comment.Author = safeUser

	comment.AuthorId = cid.(string)

	db := models.DB.Create(&comment)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	var aritcle models.Article

	models.DB.Where("article_id = ?", comment.ArticleId).First(&aritcle)

	aritcle.CommentIds = append(aritcle.CommentIds, comment.CommentId)

	models.DB.Model(&aritcle).Update("comment_ids", aritcle.CommentIds)

	if isArticle == false {
		var targetComment models.Comment
		models.DB.Where("comment_id = ?", comment.TargetId).First(&targetComment)
		targetComment.ChildrenCommentIds = append(targetComment.ChildrenCommentIds, comment.CommentId)
		models.DB.Model(&targetComment).Update("children_comment_ids", targetComment.ChildrenCommentIds)
	}

	var title string
	var targetUserId string

	if cid != user.Cid {
		if isArticle {
			title = user.Username + "评论了你的文章"
			var article models.Article
			models.DB.Where("article_id = ?", comment.TargetId).First(&article)
			targetUserId = article.Author.Cid
		} else {
			title = user.Username + "评论了你的评论"
			var targetComment models.Comment
			models.DB.Where("comment_id = ?", comment.TargetId).First(&targetComment)
			targetUserId = targetComment.Author.Cid
		}

		var noticeCode string

		noticeCode = comment.CommentId + "|" + comment.ArticleId

		SendMessage(title, comment.Content, "comment", noticeCode, &user, targetUserId, c)
	}

	c.JSON(200, models.Result{0, "success", comment.CommentId})
}

func DeleteComment(c *gin.Context) {
	commentId := c.Param("commentId")
	var comment models.Comment
	db := models.DB.Where("comment_id = ?", commentId).Where("deleted_at IS NULL").First(&comment)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	db = models.DB.Model(&comment).Where("comment_id = ?", commentId).Update("deleted_at", time.Now())
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func syncCommentInfo(comment *models.Comment, ch chan error) {
	cid := comment.Author.Cid
	var author models.User
	models.DB.Where("cid = ?", cid).First(&author)
	comment.Author = models.GetSafeUser(author)
	models.DB.Model(&comment).Where("comment_id = ?", comment.CommentId).Updates(&comment)
	ch <- nil
}

func syncCommentInfos(comments *models.Comments, ch chan error) {
	var cids []string
	var users models.Users
	for _, comment := range *comments {
		cids = append(cids, comment.Author.Cid)
	}
	models.DB.Where("cid in (?)", cids).Find(&users)
	for index, comment := range *comments {
		user := models.GetSafeUser(users[index])
		comment.Author = user
	}
	models.DB.Where("comment_id in (?)", cids).Updates(&comments)
	ch <- nil
}

func LikeComment(c *gin.Context) {
	commentId := c.Param("commentId")
	cid, _ := c.Get("cid")
	var comment models.Comment
	db := models.DB.Where("comment_id = ?", commentId).Where("deleted_at IS NULL").First(&comment)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	hasLiked := utils.ArrayIncludes[string](comment.LikeIds, commentId, func(id string) any {
		return id
	})
	if hasLiked == true {
		c.JSON(200, models.Result{10001, "您已点赞", nil})
		return
	}
	comment.LikeIds = append(comment.LikeIds, cid.(string))
	models.DB.Model(&comment).Update("like_ids", comment.LikeIds)

	var userInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid).First(&userInfo)
	userInfo.LikeCommentIds = append(userInfo.LikeCommentIds, commentId)
	models.DB.Model(&userInfo).Update("like_comment_ids", userInfo.LikeCommentIds)

	var user models.User
	models.DB.Where("cid = ?", cid).First(&user)

	if cid != comment.Author.Cid {
		title := "您有新的点赞"
		content := user.Username + "点赞了您的评论"
		noticeType := "likeComment"
		noticeId := comment.CommentId + "|" + comment.ArticleId
		SendMessage(title, content, noticeType, noticeId, &user, comment.Author.Cid, c)
	}

	c.JSON(200, models.Result{0, "success", nil})
}

func UnLikeComment(c *gin.Context) {
	commentId := c.Param("commentId")
	cid, _ := c.Get("cid")
	var comment models.Comment
	db := models.DB.Where("comment_id = ?", commentId).Where("deleted_at IS NULL").First(&comment)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	likes := utils.ArrayFilter[string](comment.LikeIds, func(id string) bool {
		return id != cid
	})
	comment.LikeIds = likes
	models.DB.Model(&comment).Update("like_ids", comment.LikeIds)

	var userInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid).First(&userInfo)
	userInfo.LikeCommentIds = utils.ArrayFilter[string](userInfo.LikeCommentIds, func(id string) bool {
		return id != commentId
	})
	models.DB.Model(&userInfo).Update("like_comment_ids", userInfo.LikeCommentIds)

	c.JSON(200, models.Result{0, "success", nil})
}

func GetMineLikeComments(c *gin.Context) {
	paginations := models.Pagination{
		PageSize: 20,
		PageNo:   1,
	}
	err := c.ShouldBindJSON(&paginations)

	if err != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	cid, _ := c.Get("cid")
	var user models.UserExtendInfo
	models.DB.Where("cid = ?", cid).First(&user)
	var comments models.Comments

	var idsInterface []interface{}
	for _, id := range user.LikeCommentIds {
		idsInterface = append(idsInterface, id)
	}

	db := models.DB.Debug().
		Table("comments").Where("deleted_at IS NULL").
		Where("comment_id in (?)", idsInterface).
		Offset((paginations.PageNo - 1) * paginations.PageSize).
		Limit(paginations.PageSize).
		Order("comment_id desc").
		Find(&comments)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	pageNo := paginations.PageNo
	pageSize := paginations.PageSize

	count := len(user.LikeCommentIds)

	pageCount := int64(count)

	list := models.GetListData[models.Comment](comments, pageNo, pageSize, pageCount)

	c.JSON(200, models.Result{0, "success", list})
}
