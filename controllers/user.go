package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-pet-family/models"
	"go-pet-family/utils"
	"gorm.io/gorm"
	"time"
)

func GetUsers(c *gin.Context) {
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
	var users models.Users
	db := models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Where("deleted_at IS NULL").Find(&users)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	var totalCount int64
	models.DB.Model(&users).Count(&totalCount)
	safeUsers := models.GetSafeUsers(users)
	list := models.GetListData[models.SafeUser](safeUsers, pageNo, pageSize, totalCount)
	c.JSON(200, models.Result{0, "success", list})
}

func GetUser(c *gin.Context) {
	cid := c.Param("cid")
	var userDetail models.UserDetail
	db := models.DB.Debug().Table("user").
		Select("*").
		Joins("LEFT JOIN user_extend_infos uei ON uei.cid = user.cid").
		Where("user.cid = ?", cid).
		Where("deleted_at IS NULL").
		First(&userDetail)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", models.GetSafeUserDetail(userDetail)})
}

func CreateUser(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if user.Email != "" && err != nil {
		// 显示自定义的错误信息
		msg := utils.GetValidMsg(err, &user)
		c.JSON(200, models.Result{Code: 10001, Message: msg})
		return
	}

	if user.Email == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
		return
	}

	if user.Email != "" {
		emailExist := checkEmailExists(user.Email, "")
		if emailExist {
			c.JSON(200, models.Result{Code: 10002, Message: "邮箱已存在"})
			return
		}
	}
	if user.Phone != "" {
		phoneExist := checkPhoneExists(user.Phone, "")
		if phoneExist {
			c.JSON(200, models.Result{Code: 10002, Message: "手机号已存在"})
			return
		}
	}

	newUUID := uuid.New()
	uuidSring := newUUID.String()
	user.Cid = "C-" + uuidSring

	user.Password = "123456"

	db := models.DB.Create(&user)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	ch := make(chan string)
	go CheckAndCreateExtendUserInfo(c, user.Cid, ch)
	<-ch

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: models.GetSafeUser(user)})
}

func CreateByEmail(ch chan string, c *gin.Context, email string) {
	var user models.User
	user.Email = email
	if user.Email == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
		return
	}
	checkEmail := utils.CheckEmail(email)
	if !checkEmail {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱格式不正确"})
	}

	newUUID := uuid.New()
	uuidSring := newUUID.String()
	user.Cid = "C-" + uuidSring

	user.Password = "123456"

	db := models.DB.Create(&user)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	chUE := make(chan string)
	go CheckAndCreateExtendUserInfo(c, user.Cid, ch)
	<-chUE
	ch <- "success"
}

func CheckAndCreateExtendUserInfo(c *gin.Context, cid string, ch chan string) {
	var userExtendInfo models.UserExtendInfo
	userExtendInfo.Cid = cid
	db := models.DB.Where("cid = ?", cid).First(&userExtendInfo)
	if db.Error == nil {
		ch <- "success"
	} else if db.Error == gorm.ErrRecordNotFound {
		db := models.DB.Create(&userExtendInfo)
		if db.Error != nil {
			// SQL执行失败，返回错误信息
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
			return
		}
		ch <- "success"
	} else {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
}

func UpdateUser(c *gin.Context) {
	cid := c.Param("cid")
	var user models.UpdateUserFields
	var oldUser models.User
	getUser := models.DB.Where("cid = ?", cid).Where("deleted_at IS NULL").First(&oldUser)
	if getUser.Error != nil {
		if getUser.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	err := c.ShouldBindJSON(&user)
	if err != nil {
		// 显示自定义的错误信息
		msg := utils.GetValidMsg(err, &user)
		c.JSON(200, models.Result{Code: 10001, Message: msg})
		return
	}
	if user.Email != "" {
		emailExist := checkEmailExists(user.Email, oldUser.Email)
		if emailExist {
			c.JSON(200, models.Result{Code: 10002, Message: "邮箱已存在"})
			return
		}
	}
	if user.Phone != "" {
		phoneExist := checkPhoneExists(user.Phone, oldUser.Phone)
		if phoneExist {
			c.JSON(200, models.Result{Code: 10002, Message: "手机号已存在"})
			return
		}
	}
	var newUser models.User
	db := models.DB.Model(&oldUser).Where("cid = ?", cid).Updates(&newUser)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	resultUser := models.User{
		Cid:      oldUser.Cid,
		Email:    newUser.Email,
		Phone:    newUser.Phone,
		Avatar:   newUser.Avatar,
		Age:      newUser.Age,
		Username: newUser.Username,
		Gender:   newUser.Gender,
		Birthday: newUser.Birthday,
		Role:     newUser.Role,
	}

	updateCh := make(chan string)
	go models.CommonUpdate[models.User](&resultUser, c, updateCh)
	<-updateCh

	syncExtendInfo := make(chan string)
	go CheckAndCreateExtendUserInfo(c, resultUser.Cid, syncExtendInfo)
	<-syncExtendInfo
	c.JSON(200, models.Result{Code: 0, Message: "success", Data: models.GetSafeUser(resultUser)})
}

func DeleteUser(c *gin.Context) {
	cid := c.Param("cid")
	fmt.Println(cid)
	db := models.DB.Model(&models.User{}).Where("cid = ?", cid).Update("deleted_at", time.Now())
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{Code: 0, Message: "success"})
}

func HardDeleteUser(c *gin.Context) {
	cid := c.Param("cid")
	db := models.DB.Model(&models.User{}).Delete(&models.User{}, "cid = ?", cid)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{Code: 0, Message: "success"})
}

func checkEmailExists(email string, exceptedEmail string) bool {
	var user models.User
	var db *gorm.DB
	if exceptedEmail != "" {
		db = models.DB.Where("email != ?", exceptedEmail).Where("email = ?", email).First(&user)
	} else {
		db = models.DB.Where("email = ?", email).First(&user)
	}
	return db.Error == nil
}

func checkPhoneExists(phone string, exceptedPhone string) bool {
	var user models.User
	var db *gorm.DB
	if exceptedPhone != "" {
		db = models.DB.Where("phone != ?", exceptedPhone).Where("phone = ?", phone).First(&user)
	} else {
		db = models.DB.Where("phone = ?", phone).First(&user)
	}
	return db.Error == nil
}

func FollowUser(c *gin.Context) {
	targetCid := c.Param("cid")
	cid, _ := c.Get("cid")

	var userExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid).First(&userExtendInfo)
	hasFollow := utils.ArrayIncludes[models.SafeUser](userExtendInfo.Follows, targetCid, func(item models.SafeUser) any {
		return item.Cid
	})
	if hasFollow {
		c.JSON(200, models.Result{Code: 10001, Message: "您已关注"})
		return
	}

	var user models.User
	db := models.DB.Where("cid = ?", targetCid).First(&user)
	safeUser := models.GetSafeUser(user)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	userExtendInfo.Follows = append(userExtendInfo.Follows, safeUser)

	updateDb := models.DB.Model(&userExtendInfo).Where("cid = ?", cid).Update("follows", userExtendInfo.Follows)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	var self models.User
	models.DB.Where("cid = ?", cid).First(&self)
	safeSelfUser := models.GetSafeUser(self)

	var targetExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", targetCid).First(&targetExtendInfo)

	targetExtendInfo.Followers = append(targetExtendInfo.Followers, safeSelfUser)

	updateTargetDb := models.DB.Model(&targetExtendInfo).Where("cid = ?", targetCid).Update("followers", targetExtendInfo.Followers)

	if updateTargetDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: user.Cid})
}

func UnFollowUser(c *gin.Context) {
	targetCid := c.Param("cid")
	cid, _ := c.Get("cid")
	var userExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", cid).First(&userExtendInfo)
	hasFollow := utils.ArrayIncludes[models.SafeUser](userExtendInfo.Follows, targetCid, func(item models.SafeUser) any {
		return item.Cid
	})
	if hasFollow == false {
		c.JSON(200, models.Result{Code: 10001, Message: "您未关注"})
		return
	}
	var user models.User
	db := models.DB.Where("cid = ?", targetCid).First(&user)
	safeUser := models.GetSafeUser(user)

	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	userExtendInfo.Follows = utils.ArrayFilter[models.SafeUser](userExtendInfo.Follows, func(item models.SafeUser) bool {
		return item.Cid != safeUser.Cid
	})

	updateDb := models.DB.Model(&userExtendInfo).Where("cid = ?", cid).Update("follows", userExtendInfo.Follows)
	if updateDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	var targetExtendInfo models.UserExtendInfo
	models.DB.Where("cid = ?", targetCid).First(&targetExtendInfo)

	targetExtendInfo.Followers = utils.ArrayFilter[models.SafeUser](targetExtendInfo.Followers, func(item models.SafeUser) bool {
		return item.Cid != cid
	})

	updateTargetDb := models.DB.Model(&targetExtendInfo).Where("cid = ?", targetCid).Update("followers", targetExtendInfo.Followers)

	if updateTargetDb.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: nil})
}
