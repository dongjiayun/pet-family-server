package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
)
import "go-pet-family/models"
import "go-pet-family/utils"

func GetUsers(c *gin.Context) {
	var pagination models.Pagination
	err := c.ShouldBindJSON(&pagination)
	if err != nil {
		c.JSON(200, gin.H{"msg": err.Error()})
		return
	}
	pageNo := pagination.PageNo
	pageSize := pagination.PageSize
	var users models.Users
	models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&users)
	c.JSON(200, models.Result{200, "success", models.GetSafeUsers(users)})
}

func CreateUser(c *gin.Context) {
	var user models.User
	params := utils.BindJson(c, &user)
	fmt.Println(params)
	//models.DB.Create(&user)
}
