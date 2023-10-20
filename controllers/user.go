package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-pet-family/models"
	"go-pet-family/utils"
)

func GetUsers(c *gin.Context) {
	var pagination models.Pagination
	err := c.ShouldBindJSON(&pagination)
	if err != nil {
		c.JSON(10001, gin.H{"msg": err.Error()})
		return
	}
	pageNo := pagination.PageNo
	pageSize := pagination.PageSize
	var users models.Users
	models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&users)
	c.JSON(0, models.Result{200, "success", models.GetSafeUsers(users)})
}

func GetUser(c *gin.Context) {
	cid := c.Param("cid")
	var user models.User
	models.DB.Where("cid = ?", cid).Find(&user)
	c.JSON(0, models.Result{200, "success", models.GetSafeUser(user)})
}

func CreateUser(c *gin.Context) {
	var user models.User
	params := utils.BindJson(c, &user)
	fmt.Println(params)
	//models.DB.Create(&user)
}
