package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
	"time"
)

type GetCommentsReq struct {
	models.Pagination
	TargetId string `json:"targetId"`
	Sync     bool   `json:"sync"`
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

	fmt.Println("commentsReq", commentsReq)

	var comments models.Comments

	db := models.DB.Debug().Where("target_id = ?", targetId).Where("deleted_at IS NULL").Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&comments)

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
	models.DB.Model(&comments).Count(&count)

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

	ch := make(chan string)
	go models.CommonCreate[models.Comment](&comment, ch)
	<-ch

	db := models.DB.Create(&comment)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
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
