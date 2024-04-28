package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
	"time"
)

type ResumeReq struct {
	models.Pagination
	Language string `json:"language"`
	IsLatest bool   `json:"isLatest"`
}

func GetResumes(c *gin.Context) {
	resumeReq := ResumeReq{
		Pagination: models.Pagination{
			PageSize: 20,
			PageNo:   1,
		},
	}
	err := c.ShouldBindJSON(&resumeReq)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	pageNo := resumeReq.PageNo
	pageSize := resumeReq.PageSize
	language := resumeReq.Language
	IsLatest := resumeReq.IsLatest
	var resumes models.Resumes

	var count int64

	db := models.DB.Model(&models.Resume{}).
		Where("deleted_at IS NULL")

	if language != "" {
		db = db.Where("language = ?", language)
	}

	if IsLatest {
		var en models.Resume
		var zh models.Resume

		// 查询英文简历
		models.DB.Model(&models.Resume{}).Where("language = 'en'").Order("updated_at desc").First(&en)
		if en.ResumeId != "" {
			resumes = append(resumes, en)
		}

		// 查询中文简历
		models.DB.Model(&models.Resume{}).Where("language = 'zh'").Order("updated_at desc").First(&zh)
		if zh.ResumeId != "" {
			resumes = append(resumes, zh)
		}

		count = int64(len(resumes))
	} else {
		db.Limit(pageSize).Offset((pageNo - 1) * pageSize).
			Order("id desc").Count(&count).Find(&resumes)
	}

	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10001, Message: db.Error.Error()})
		return
	}

	list := models.GetListData[models.Resume](resumes, pageNo, pageSize, count)

	c.JSON(200, models.Result{0, "success", list})
}

func GetResume(c *gin.Context) {
	id := c.Param("resumeId")
	var resume models.Resume
	db := models.DB.Where("resume_id = ?", id).First(&resume)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10001, Message: db.Error.Error()})
		return
	}
	c.JSON(200, models.Result{0, "success", resume})
}

func CreateResume(c *gin.Context) {
	var resume models.Resume
	err := c.ShouldBindJSON(&resume)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	uuid := uuid.New()
	uuidStr := uuid.String()
	resume.ResumeId = "Resume-" + uuidStr

	models.CommonCreate[models.Resume](&resume)

	db := models.DB.Create(&resume)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10001, Message: db.Error.Error()})
	}
	c.JSON(200, models.Result{0, "success", &resume.ResumeId})
}

func UpdateResume(c *gin.Context) {
	var resumeReq models.Resume
	err := c.ShouldBindJSON(&resumeReq)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	resumeId := resumeReq.ResumeId
	resume := models.Resume{}
	db := models.DB.Where("resume_id = ?", resumeId).Where("deleted_at IS NULL").First(&resume)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	update := models.Resume{
		Language: resumeReq.Language,
		Content:  resumeReq.Content,
		Version:  resume.Version + 1,
	}

	uploadDb := models.DB.
		Model(&update).
		Where("resume_id = ?", resumeId).
		Updates(&update)

	if uploadDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", &resume.ResumeId})
}

func DeleteResume(c *gin.Context) {
	id := c.Param("resumeId")
	var resume models.Resume
	db := models.DB.Where("resume_id = ?", id).Where("deleted_at IS NULL").First(&resume)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	models.DB.Model(&resume).Update("deleted_at", time.Now())
	c.JSON(200, models.Result{0, "success", nil})
}
