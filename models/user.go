package models

import (
	"database/sql/driver"
	"errors"
	"github.com/goccy/go-json"
)

type User struct {
	Model
	Username  string `json:"username"`
	Password  string `json:"-"`
	Email     string `json:"email" binding:"email" msg:"请输入正确的邮箱地址" gorm:"index"`
	Phone     string `json:"phone" binding:"phone" msg:"请输入正确的手机号" gorm:"index"`
	Cid       string `json:"cid" gorm:"index"`
	Unionid   string `json:"unionid" gorm:"index"`
	Openid    string `json:"openid" gorm:"index"`
	IsDeleted bool   `json:"-"`
	Gender    string `json:"gender"`
	Birthday  string `json:"birthday"`
	Avatar    File   `json:"avatar" gorm:"type:json"`
	AvatarId  string `json:"-" gorm:"type:varchar(255)"`
	Age       int    `json:"age"`
	Role      string `json:"role"`
}

type UserExtendInfo struct {
	Id             uint   `json:"-" gorm:"primary_key"`
	Cid            string `json:"cid" gorm:"index"`
	CommentIds     Ids    `json:"commentIds" gorm:"type:json"`
	LikeArticleIds Ids    `json:"likesArticleIds" gorm:"type:json"`
	LikeCommentIds Ids    `json:"likesCommentIds" gorm:"type:json"`
	CollectIds     Ids    `json:"collectIds" gorm:"type:json"`
	FollowIds      Ids    `json:"followIds" gorm:"type:json"`
	FollowerIds    Ids    `json:"followerIds" gorm:"type:json"`
}

type Users []User

func (user User) TableName() string {
	return "user"
}

type SafeUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Cid      string `json:"cid"`
	Gender   string `json:"gender"`
	Birthday string `json:"birthday"`
	Avatar   File   `json:"avatar" gorm:"type:json"`
	Age      int    `json:"age"`
	Role     string `json:"role"`
}

func (safeUser *SafeUser) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*safeUser = SafeUser{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, safeUser)
}

func (safeUser SafeUser) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(safeUser)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type SafeUsers []SafeUser

func (users *SafeUsers) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*users = []SafeUser{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, users)
}

func (users SafeUsers) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(users)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

func GetSafeUser(user User) SafeUser {
	maskedPhone := maskPhoneNumber(user.Phone)
	safeUser := SafeUser{
		Username: user.Username,
		Email:    user.Email,
		Phone:    maskedPhone,
		Cid:      user.Cid,
		Gender:   user.Gender,
		Birthday: user.Birthday,
		Avatar:   user.Avatar,
		Age:      user.Age,
		Role:     user.Role,
	}
	return safeUser
}

func GetSafeUsers(users []User) []SafeUser {
	var safeUsers []SafeUser

	// 遍历原始用户列表，将其转换为SafeUser结构体并添加到safeUsers切片中
	for _, user := range users {
		safeUser := SafeUser{
			Username: user.Username,
			Email:    user.Email,
			Phone:    maskPhoneNumber(user.Phone),
			Cid:      user.Cid,
			Gender:   user.Gender,
			Birthday: user.Birthday,
			Avatar:   user.Avatar,
			Age:      user.Age,
			Role:     user.Role,
		}
		safeUsers = append(safeUsers, safeUser)
	}

	return safeUsers
}

type UpdateUserFields struct {
	Cid      string  `json:"cid" binding:"required"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Gender   *string `json:"gender"`
	Birthday *string `json:"birthday"`
	Avatar   *File   `json:"avatar" gorm:"foreignKey:FileId;type:varchar(255)"`
	Age      *int    `json:"age"`
	Role     *string `json:"role"`
}

type AuthUser struct {
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Otp          string `json:"otp"`
	LoginType    string `json:"loginType"`
	Ticket       string `json:"ticket"`
	RefreshToken string `json:"refreshToken"`
	JsCode       string `json:"jsCode"`
}

type AuthOtp struct {
	Code    string `json:"code"`
	Account string `json:"account"`
	Ticket  string `json:"ticket"`
}

type UserDetail struct {
	User
	CommentIds     Ids `json:"commentIds"`
	LikeArticleIds Ids `json:"likesArticleIds"`
	LikeCommentIds Ids `json:"likesCommentIds"`
	CollectIds     Ids `json:"collectIds""`
	FollowIds      Ids `json:"followIds"`
	FollowerIds    Ids `json:"followerIds"`
}

type UserDetailList []UserDetail

type SafeUserDetail struct {
	SafeUser
	CommentIds     Ids `json:"commentIds"`
	LikeArticleIds Ids `json:"likesArticleIds"`
	LikeCommentIds Ids `json:"likesCommentIds"`
	CollectIds     Ids `json:"collectIds""`
	FollowIds      Ids `json:"followIds"`
	FollowerIds    Ids `json:"followerIds"`
}

type SafeUserDetailList []SafeUserDetail

func GetSafeUserDetail(user UserDetail) SafeUserDetail {
	safeUser := GetSafeUser(user.User)
	safeUserDetail := SafeUserDetail{
		SafeUser: safeUser,
	}
	safeUserDetail.CommentIds = user.CommentIds
	safeUserDetail.LikeArticleIds = user.LikeArticleIds
	safeUserDetail.LikeCommentIds = user.LikeCommentIds
	safeUserDetail.CollectIds = user.CollectIds
	safeUserDetail.FollowIds = user.FollowIds
	safeUserDetail.FollowerIds = user.FollowerIds
	return safeUserDetail
}

func GetSafeUserDetailList(userDetailList UserDetailList) SafeUserDetailList {
	safeUserDetailList := SafeUserDetailList{}
	for _, user := range userDetailList {
		safeUserDetail := GetSafeUserDetail(user)
		safeUserDetailList = append(safeUserDetailList, safeUserDetail)
	}
	return safeUserDetailList
}
