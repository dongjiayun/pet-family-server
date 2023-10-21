package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
	"go-pet-family/utils"
)

func GetUsers(c *gin.Context) {
	var pagination models.Pagination
	err := c.ShouldBindJSON(&pagination)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	pageNo := pagination.PageNo
	pageSize := pagination.PageSize
	var users models.Users
	db := models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&users)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(0, models.Result{200, "success", models.GetSafeUsers(users)})
}

func GetUser(c *gin.Context) {
	cid := c.Param("cid")
	var user models.User
	db := models.DB.Where("cid = ?", cid).Find(&user)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(0, models.Result{200, "success", models.GetSafeUser(user)})
}

func CreateUser(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		// 显示自定义的错误信息
		msg := utils.GetValidMsg(err, &user)
		c.JSON(200, models.Result{Code: 10001, Message: msg})
		return
	}
	if user.Email != "" {
		emailExist := checkEmailExists(user.Email)
		if emailExist {
			c.JSON(200, models.Result{Code: 10002, Message: "邮箱已存在"})
			return
		}
	}
	if user.Phone != "" {
		phoneExist := checkPhoneExists(user.Phone)
		if phoneExist {
			c.JSON(200, models.Result{Code: 10002, Message: "手机号已存在"})
			return
		}
	}

	newUUID := uuid.New()
	uuidSring := newUUID.String()
	user.Cid = "C" + uuidSring

	db := models.DB.Create(&user)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(0, models.Result{Code: 200, Message: "success", Data: models.GetSafeUser(user)})
}

func checkEmailExists(email string) bool {
	var user models.User
	db := models.DB.Where("email = ?", email).First(&user)
	return db.Error == nil
}

func checkPhoneExists(phone string) bool {
	var user models.User
	db := models.DB.Where("phone = ?", phone).First(&user)
	return db.Error == nil
}
