package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-gomail/gomail"
	"github.com/google/uuid"
	"go-pet-family/config"
	"go-pet-family/models"
	"go-pet-family/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type MyClaims struct {
	Cid string `json:"cid"`
	jwt.StandardClaims
}

const TokenExpireDuration = time.Hour * 24

var Secret = []byte(config.Secret)

func SignIn(c *gin.Context) {
	var user models.AuthUser
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	if user.LoginType == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "请选择登录方式"})
	}
	switch user.LoginType {
	case "phone":
		if user.Phone == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "手机号不能为空"})
			return
		}
		if user.Otp == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "验证码不能为空"})
		}
		phoneExist := checkPhoneExists(user.Phone, "")
		if phoneExist {

		} else {
			c.JSON(200, models.Result{Code: 10001, Message: "手机号不存在"})
		}
	case "email":
		if user.Email == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
			return
		}
		emailExist := checkEmailExists(user.Email, "")
		if emailExist {
			if user.Otp == "" {
				c.JSON(200, models.Result{Code: 10001, Message: "验证码不能为空"})
				return
			}

			optCache := models.RedisClient.Get(context.Background(), user.Email)

			if optCache.Val() != "" {
				var cache models.AuthOtp
				json.Unmarshal([]byte(optCache.Val()), &cache)
				if cache.Code == user.Otp && cache.Ticket == user.Ticket {
					generateToken(c, user.Email, "email")
					return
				} else {
					c.JSON(200, models.Result{Code: 10001, Message: "验证码错误"})
					return
				}
			} else {
				c.JSON(200, models.Result{Code: 10001, Message: "请发送验证码"})
			}
		} else {
			ch := make(chan string)
			go CreateByEmail(ch, c, user.Email)
			result := <-ch
			if result == "success" {
				generateToken(c, user.Email, "email")
				return
			}
		}
	case "emailWithPassword":
		if user.Email == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
			return
		}
		if user.Password == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "密码不能为空"})
			return
		}
		emailExist := checkEmailExists(user.Email, "")
		if emailExist {
			var resultUser models.User
			db := models.DB.Model(&models.User{}).Where("email = ?", user.Email).First(&resultUser)
			if db.Error != nil {
				// SQL执行失败，返回错误信息
				c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
				return
			}
			if resultUser.Password == user.Password {
				generateToken(c, user.Email, "email")
			} else {
				c.JSON(200, models.Result{Code: 10001, Message: "密码错误"})
			}
		} else {
			c.JSON(200, models.Result{Code: 10001, Message: "邮箱不存在"})
		}
	}

	if user.Email == "" && user.Password == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱或手机号不能为空"})
		return
	}
}

func generateToken(c *gin.Context, account string, loginType string) {
	var resultUser models.User

	if loginType == "email" {
		db := models.DB.Model(&resultUser).Where("email = ?", account).First(&resultUser)
		if db.Error != nil {
			// SQL执行失败，返回错误信息
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		}
	}

	token, _ := GenToken(resultUser.Cid)

	redisClient := models.RedisClient

	redisClient.Del(context.Background(), account)

	type Result struct {
		models.User
		Token string `json:"token"`
	}

	result := Result{
		User:  resultUser,
		Token: token,
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: result})
}

func SendEmailOtp(c *gin.Context) {
	type OtpCode struct {
		Email string `json:"email" binding:"email" msg:"请输入正确的邮箱地址" gorm:"index"`
	}
	var otpCode OtpCode
	err := c.ShouldBindJSON(&otpCode)
	if otpCode.Email != "" && err != nil {
		// 显示自定义的错误信息
		msg := utils.GetValidMsg(err, &otpCode)
		c.JSON(200, models.Result{Code: 10001, Message: msg})
		return
	}
	if otpCode.Email == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
		return
	}
	emailExist := checkEmailExists(otpCode.Email, "")
	if !emailExist {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不存在"})
		return
	}

	optCache := models.RedisClient.Get(context.Background(), otpCode.Email)

	if optCache.Val() != "" {
		var cache models.AuthOtp
		json.Unmarshal([]byte(optCache.Val()), &cache)
		c.JSON(200, models.Result{
			Code:    10001,
			Message: "验证码已发送,请勿重复发送",
			Data:    cache.Ticket,
		})
		return
	}

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(900000) + 100000
	randomNumberStr := strconv.Itoa(randomNumber)

	to := otpCode.Email
	subject := "邮箱验证码"
	message := "验证码：" + randomNumberStr

	smtpHost := config.SmtpHost
	smtpPort := config.SmtpPort
	smtpUser := config.SmtpUser
	smtpPassword := config.SmtpPassword

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)   // 发件人邮箱
	m.SetHeader("To", to)           // 收件人邮箱
	m.SetHeader("Subject", subject) // 邮件主题
	m.SetBody("text/html", message) // 邮件内容

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ticket := uuid.New()
	ticketString := ticket.String()

	authOtp := models.AuthOtp{
		Code:    randomNumberStr,
		Account: otpCode.Email,
		Ticket:  ticketString,
	}

	authOtpJSON, _ := json.Marshal(authOtp)

	redisClient := models.RedisClient

	msg := redisClient.Set(context.Background(), otpCode.Email, authOtpJSON, 5*time.Minute)

	if msg != nil {
		fmt.Println(msg)
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: ticketString})
}

func SignOut(c *gin.Context) {

}

func SignUp(c *gin.Context) {

}

// GenToken 生成JWT
func GenToken(Cid string) (string, error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		Cid, // 自定义字段
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(), // 过期时间
			Issuer:    "pet-family",                               // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(Secret)
}

type TokenClaims struct {
	CID string
}

func CheckToken(c *gin.Context) (*TokenClaims, error) {
	tokenString := c.GetHeader("Authorization")
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 获取用户的CID（假设你的 MyClaims 结构体中有一个 CID 字段）
	cid := claims.Cid

	fmt.Println(claims)

	// 返回 TokenClaims 结构体，包含 MyClaims 和 CID
	return &TokenClaims{
		CID: cid,
	}, nil
}
