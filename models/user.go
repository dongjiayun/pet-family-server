package models

func init() {
	DB.AutoMigrate(&User{})
}

type User struct {
	Username  string `json:"username"`
	Password  string `json:"-"`
	Email     string `json:"email" binding:"email" msg:"请输入正确的邮箱地址" gorm:"index"`
	Phone     string `json:"phone" binding:"phone" msg:"请输入正确的手机号" gorm:"index"`
	Cid       string `json:"cid"`
	IsDeleted bool   `json:"-"`
	Id        int    `json:"-" gorm:"primary_key"`
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
}

func GetSafeUser(user User) SafeUser {
	maskedPhone := maskPhoneNumber(user.Phone)
	safeUser := SafeUser{
		Username: user.Username,
		Email:    user.Email,
		Phone:    maskedPhone,
		Cid:      user.Cid,
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
		}
		safeUsers = append(safeUsers, safeUser)
	}

	return safeUsers
}
