package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
	selfUtils "go-pet-family/utils"
	gormUtils "gorm.io/gorm/utils"
	"time"
)

func GetMessage(c *gin.Context) {
	pagination := models.Pagination{
		PageSize: 20,
		PageNo:   1,
	}
	c.ShouldBindJSON(&pagination)
	var Messages []models.Message
	db := models.DB.
		Limit(pagination.PageSize).Offset((pagination.PageNo - 1) * pagination.PageSize).
		Order("id desc").
		Where("deleted_at IS NULL").
		Find(&Messages)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	var total int64

	db = models.DB.Model(&models.Message{}).Where("deleted_at IS NULL").Count(&total)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	list := models.GetListData[models.Message](Messages, pagination.PageNo, pagination.PageSize, total)
	c.JSON(200, models.Result{0, "success", list})
}

type CreateMessageReq struct {
	Content string `json:"content"`
}

func CreateMessage(c *gin.Context) {
	var req CreateMessageReq
	c.ShouldBindJSON(&req)
	var message models.Message
	message.Content = req.Content

	message.MessageId = uuid.New().String()

	cid, _ := c.Get("cid")
	if cid != nil {
		message.AuthorId = cid.(string)
		var user models.User
		models.DB.Where("cid = ?", message.AuthorId).First(&user)
		message.Author = models.GetSafeUser(user)
	}

	models.DB.Create(&message)

	c.JSON(200, models.Result{0, "success", message})
}

func DeleteMessage(c *gin.Context) {
	cid, _ := c.Get("cid")
	messageId := c.Param("messageId")
	var message models.Message
	models.DB.Where("message_id = ?", messageId).First(&message)

	var ch = make(chan string)
	go CheckSelfOrAdmin(c, cid.(string), ch)
	msg := <-ch

	if msg != "success" {
		c.JSON(200, models.Result{Code: 10001, Message: msg})
		return
	}

	db := models.DB.Where("message_id = ?", messageId).Update("deleted_at", time.Now())
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	c.JSON(200, models.Result{0, "success", nil})
}

func LikeMessage(c *gin.Context) {
	cid, _ := c.Get("cid")
	messageId := c.Param("messageId")
	var message models.Message
	db := models.DB.Where("message_id = ?", messageId).First(&message)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	likes := message.LikeIds
	if gormUtils.Contains(likes, cid.(string)) {
		c.JSON(200, models.Result{Code: 10003, Message: "已点赞"})
		return
	}
	message.LikeIds = append(likes, cid.(string))
	db = models.DB.Model(&message).Where("message_id = ?", messageId).Updates(&message)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func UnLikeMessage(c *gin.Context) {
	cid, _ := c.Get("cid")
	messageId := c.Param("messageId")
	var message models.Message
	db := models.DB.Where("message_id = ?", messageId).First(&message)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	likes := message.LikeIds
	if !gormUtils.Contains(likes, cid.(string)) {
		c.JSON(200, models.Result{Code: 10003, Message: "未点赞"})
		return
	}
	message.LikeIds = selfUtils.ArrayFilter[string](message.LikeIds, func(id string) bool {
		return id != cid
	})
	db = models.DB.Model(&message).Where("message_id = ?", messageId).Updates(&message)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", nil})
}
