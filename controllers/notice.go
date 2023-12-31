package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	db := models.DB.Where("target_cid = ?", cid).
		Where("deleted_at IS NULL").
		Limit(pageSize).Offset((pageNo - 1) * pageSize).
		Order("is_readed asc").Order("id desc").Find(&notices)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
	}
	var count int64
	models.DB.Model(&notices).Where("target_cid = ?", cid).Where("deleted_at IS NULL").Count(&count)

	list := models.GetListData[models.Notice](notices, pageNo, pageSize, count)

	c.JSON(200, models.Result{0, "success", list})
}

func ReadNotice(c *gin.Context) {
	noticeId := c.Param("noticeId")
	cid, _ := c.Get("cid")
	var notice models.Notice
	db := models.DB.Where("notice_id = ?", noticeId).Where("target_cid = ?", cid).First(&notice)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 0, Message: "success"})
			return
		}
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
	}
	readDb := models.DB.Model(&notice).Where("notice_id = ?", noticeId).Update("is_readed", true)
	if readDb.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func ReadAllNotices(c *gin.Context) {
	var notices models.Notices
	cid, _ := c.Get("cid")
	models.DB.Where("target_cid = ?", cid).Where("deleted_at IS NULL AND is_readed = ?", false).Limit(20).Offset(0).Order("id desc").Find(&notices)
	for _, notice := range notices {
		readDb := models.DB.Model(&notice).Where("notice_id = ?", notice.NoticeId).Update("is_readed", true)
		if readDb.Error != nil {
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		}
	}
	c.JSON(200, models.Result{0, "success", nil})
}

func SendArticleMessageToAllFollows(article *models.Article, c *gin.Context) {
	var userExtendInfo models.UserExtendInfo
	var user models.User
	authorCid := article.Author.Cid
	models.DB.Where("cid = ?", authorCid).Find(&userExtendInfo)
	models.DB.Where("cid = ?", authorCid).Find(&user)
	followers := userExtendInfo.FollowerIds
	for _, follower := range followers {
		var notice models.Notice
		notice.NoticeType = "article"
		notice.NoticeCode = article.ArticleId
		notice.Title = "您关注的" + user.Username + "发表了一篇文章"
		notice.Content = article.Title
		notice.Avatar = user.Avatar
		notice.TargetCid = follower

		uuid := uuid.New()
		uuidStr := uuid.String()
		notice.NoticeId = "notice-" + uuidStr

		models.CommonCreate[models.Notice](&notice)

		db := models.DB.Create(&notice)
		if db.Error != nil {
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		}
	}
}

func SendMessage(title string, content string, noticeType string, noticeCode string, user *models.User, targetCid string, c *gin.Context) {
	var notice models.Notice
	if user == nil {
		cid, _ := c.Get("cid")
		models.DB.Where("cid = ?", cid).Find(&user)
	}
	notice.NoticeType = noticeType
	notice.NoticeCode = noticeCode
	notice.Title = title
	notice.Content = content
	notice.Avatar = user.Avatar
	notice.TargetCid = targetCid

	uuid := uuid.New()
	uuidStr := uuid.String()
	notice.NoticeId = "notice-" + uuidStr

	models.CommonCreate[models.Notice](&notice)

	db := models.DB.Create(&notice)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
	}
}

func GetNoticeAmount(c *gin.Context) {
	var notices models.Notices
	cid, _ := c.Get("cid")
	var amount int64
	db := models.DB.Model(&notices).
		Where("target_cid = ? and is_readed = 0", cid).Where("deleted_at IS NULL").Count(&amount)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
	}
	c.JSON(200, models.Result{0, "success", amount})
}
