package controllers

import (
	"github.com/gin-gonic/gin"
	"go-pet-family/models"
)

type GetNoticesReq struct {
	models.Pagination
}

func GetNotices(c *gin.Context) {
	noticeReq := GetNoticesReq{
		Pagination: models.Pagination{
			PageSize: 20,
			PageNo:   1,
		},
	}
	err := c.ShouldBindJSON(&noticeReq)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	pageNo := noticeReq.PageNo
	pageSize := noticeReq.PageSize
	cid, _ := c.Get("cid")
	var notices models.Notices
	db := models.DB.Where("owner = ?", cid).Where("deleted_at IS NULL").Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&notices)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
	}
	var count int64
	models.DB.Model(&notices).Count(&count)

	list := models.GetListData[models.Notice](notices, pageNo, pageSize, count)

	c.JSON(200, models.Result{0, "success", list})
}

func SendArticleMessageToAllFollows(article *models.Article) {
	var users models.Users
	authorCid := article.Author.Cid
	models.DB.Table("user_extend_infos").Where("cid = ?", authorCid).Find(&users)
}
