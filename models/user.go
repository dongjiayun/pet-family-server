package models

func init() {
	DB.AutoMigrate(&User{})
}

type User struct {
	Username   string `json:"username"`
	Password   string `json:"-"`
	Email      string `json:"email" binding:"email" msg:"请输入正确的邮箱地址" gorm:"index"`
	Phone      string `json:"phone" binding:"phone" msg:"请输入正确的手机号" gorm:"index"`
	Cid        string `json:"cid" gorm:"index"`
	IsDeleted  bool   `json:"-"`
	Id         int    `json:"-" gorm:"primary_key"`
	Gender     string `json:"gender"`
	Birthday   string `json:"birthday"`
	Avatar     string `json:"avatar"`
	Age        int    `json:"age"`
	CreateTime string `json:"-"`
	UpdateTime string `json:"-"`
	Role       string `json:"role"`
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
	Avatar   string `json:"avatar"`
	Age      int    `json:"age"`
	Role     string `json:"role"`
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
	Username string `json:"username"`
	Email    string `json:"email" binding:"email" msg:"请输入正确的邮箱地址" gorm:"index"`
	Phone    string `json:"phone" binding:"phone" msg:"请输入正确的手机号" gorm:"index"`
	Gender   string `json:"gender"`
	Birthday string `json:"birthday"`
	Avatar   string `json:"avatar"`
	Age      int    `json:"age"`
	Role     string `json:"role"`
}
