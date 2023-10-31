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
	db := models.DB.Where("owner = ?", cid).Where("deleted_at IS NULL").Limit(pageSize).Offset((pageNo - 1) * pageSize).Order("id desc").Find(&notices)
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

func SendArticleMessageToAllFollows(article *models.Article, c *gin.Context) {
	var userExtendInfo models.UserExtendInfo
	var user models.User
	authorCid := article.Author.Cid
	models.DB.Where("cid = ?", authorCid).Find(&userExtendInfo)
	models.DB.Where("cid = ?", authorCid).Find(&user)
	followers := userExtendInfo.Followers
	for _, follower := range followers {
		var notice models.Notice
		notice.NoticeType = "article"
		notice.NoticeCode = article.ArticleId
		notice.Title = "您关注的" + user.Username + "发表了一篇文章"
		notice.Content = article.Title
		notice.Avatar = user.Avatar
		notice.TargetCid = follower.Cid

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

func SendFollowMessageToAuthor(user *models.User, c *gin.Context) {
	var notice models.Notice
	notice.NoticeType = "follow"
	notice.NoticeCode = user.Cid
	notice.Title = "您关注的" + user.Username + "发表了一篇文章"
	notice.Content = "您关注的" + user.Username + "发表了一篇文章"
	notice.Avatar = user.Avatar
	notice.TargetCid = user.Cid

}
