package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
)

func GetTags(c *gin.Context) {
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
	var tags []models.Tag
	db := models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Where("deleted_at IS NULL").Find(&tags)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	list := models.GetListData[models.Tag](tags, pageNo, pageSize, db.RowsAffected)
	c.JSON(200, models.Result{0, "success", list})
}

func GetTag(c *gin.Context) {
	tagId := c.Param("tagId")
	var tag models.Tag
	db := models.DB.Where("tag_id = ?", tagId).Where("deleted_at IS NULL").First(&tag)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", tag})
}

func CreateTag(c *gin.Context) {
	var tag models.Tag
	err := c.ShouldBindJSON(&tag)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}

	newUUID := uuid.New()
	uuidSring := newUUID.String()
	tag.TagId = "Tag-" + uuidSring
	db := models.DB.Create(&tag)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", tag})
}

func UpdateTag(c *gin.Context) {
	tagId := c.Param("tagId")
	var oldTag models.Tag
	db := models.DB.Where("tag_id = ?", tagId).Where("deleted_at IS NULL").First(&oldTag)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	var tag models.Tag
	tag.TagId = tagId
	err := c.ShouldBindJSON(&tag)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	db = models.DB.Model(&oldTag).Where("tag_id = ?", tagId).Updates(&tag)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", tag})
}
