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
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type MyClaims struct {
	Cid       string `json:"cid"`
	LoginType string `json:"login_type"`
	jwt.StandardClaims
}

const TokenExpireDuration = config.TokenExpireDuration

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

			if user.Ticket == "" {
				c.JSON(200, models.Result{Code: 10001, Message: "ticket不能为空"})
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
	case "wechat":
		client := &http.Client{}

		appid := config.MiniprogramAppid

		secret := config.MiniprogramSecret

		js_code := user.JsCode

		params := url.Values{}
		params.Set("appid", appid)
		params.Set("secret", secret)
		params.Set("js_code", js_code)
		params.Set("grant_type", "authorization_code")

		queryString := params.Encode()

		req, err := http.NewRequest("GET", "https://api.weixin.qq.com/sns/jscode2session?"+queryString, nil)

		req.Header.Add("Content-Type", "application/json")

		if err != nil {
			c.JSON(200, models.Result{Code: 10001, Message: "internal server error"})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(200, models.Result{Code: 10001, Message: "internal server error"})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(200, models.Result{Code: 10001, Message: "internal server error"})
			return
		}

		type Resp struct {
			Openid     string `json:"openid"`
			SessionKey string `json:"session_key"`
			Unionid    string `json:"unionid"`
		}

		var data Resp

		err = json.Unmarshal(body, &data)

		openId := data.Openid
		unionId := data.Unionid

		openidExists := checkOpenidExists(openId)

		if openidExists {
			generateToken(c, openId, "wechat")
		} else {
			ch := make(chan string)
			go CreateByOpenid(ch, c, openId, unionId)
			result := <-ch
			if result == "success" {
				generateToken(c, openId, "wechat")
			}
		}
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
	} else if loginType == "wechat" {
		db := models.DB.Model(&resultUser).Where("openid = ?", account).First(&resultUser)
		if db.Error != nil {
			// SQL执行失败，返回错误信息
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		}
	}

	token, _ := GenToken(resultUser.Cid, loginType)

	refreshToken, _ := GenRefreshToken(resultUser.Cid, loginType)

	type Result struct {
		models.SafeUser
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	}

	result := Result{
		SafeUser:     models.GetSafeUser(resultUser),
		Token:        token,
		RefreshToken: refreshToken,
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

	msg := redisClient.Set(context.Background(), otpCode.Email, authOtpJSON, 1*time.Minute)

	if msg != nil {
		fmt.Println(msg)
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: ticketString})
}

type WastedToken struct {
	Token      string `json:"token"`
	CreateTime int    `json:"create_time"`
}

func SignOut(c *gin.Context) {
	token := c.GetHeader("Authorization")
	handleWasteToken(token)
	c.JSON(200, models.Result{Code: 0, Message: "success"})
}

func handleWasteToken(token string) {
	redisClient := models.RedisClient
	blackList := redisClient.Get(context.Background(), "blackList")
	blackListValue := blackList.Val()
	var _blackList []WastedToken
	if blackListValue != "" {
		redisClient.Set(context.Background(), "blackList", "", 0)
	}
	err := json.Unmarshal([]byte(blackListValue), &_blackList)
	if err != nil {
		// 处理解析错误
		fmt.Println("解析JSON出错:", err)
		// 返回错误或者其他逻辑处理
	}
	wastedToken := WastedToken{
		Token:      token,
		CreateTime: int(time.Now().Unix()),
	}
	_blackList = append(_blackList, wastedToken)
	__blackList, _ := json.Marshal(_blackList)
	redisClient.Set(context.Background(), "blackList", __blackList, 0)
}

func RefreshToken(c *gin.Context) {
	refreshToken, _ := CheckRefreshToken(c)
	if refreshToken == nil {
		c.JSON(200, models.Result{Code: 10001, Message: "无效的RefreshToken"})
		return
	}
	cid := refreshToken.Cid
	fmt.Println(cid)
	loginType := refreshToken.LoginType
	var user models.User
	db := models.DB.Model(&models.User{}).Where("cid = ?", cid).First(&user)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	generateToken(c, user.Email, loginType)
}

func CheckRefreshToken(c *gin.Context) (*TokenClaims, error) {
	refreshToken := c.GetHeader("Refresh-Token")
	if refreshToken == "" {
		return nil, errors.New("无效的RefreshToken")
	}

	_, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(refreshToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyClaims)

	cid := claims.Cid
	loginType := claims.LoginType

	fmt.Println(cid, loginType)

	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return &TokenClaims{
		Cid:       cid,
		LoginType: loginType,
	}, nil
}

// GenToken 生成JWT
func GenToken(Cid string, LoginType string) (string, error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		Cid, // 自定义字段
		LoginType,
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

func GenRefreshToken(Cid string, LoginType string) (string, error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		Cid, // 自定义字段
		LoginType,
		jwt.StandardClaims{
			Issuer: "pet-family", // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(Secret)
}

type TokenClaims struct {
	Cid       string
	LoginType string
}

func CheckToken(c *gin.Context) (*TokenClaims, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return nil, errors.New("无效的token")
	}

	redisClient := models.RedisClient
	blackList := redisClient.Get(context.Background(), "blackList")
	blackListValue := blackList.Val()
	var _blackList []WastedToken
	if blackListValue != "" {
		err := json.Unmarshal([]byte(blackListValue), &_blackList)
		if err != nil {
			// 处理解析错误
			fmt.Println("解析JSON出错:", err)
			// 返回错误或者其他逻辑处理
		}
		var isWasted bool
		var newBlackList []WastedToken
		for _, wasted := range _blackList {
			if wasted.Token == tokenString {
				isWasted = true
			}
			nowTime := int(time.Now().Unix())
			expiredTime := int(time.Unix(int64(wasted.CreateTime), 0).Add(TokenExpireDuration).Unix())
			if nowTime < expiredTime {
				newBlackList = append(newBlackList, wasted)
			}
		}
		__blackList, _ := json.Marshal(newBlackList)
		redisClient.Set(context.Background(), "blackList", __blackList, 0)
		if isWasted {
			return nil, errors.New("token已失效")
		}
	}

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

	// 返回 TokenClaims 结构体，包含 MyClaims 和 CID
	return &TokenClaims{
		Cid: cid,
	}, nil
}

func CheckSelfOrAdmin(c *gin.Context, cid string, ch chan bool) {
	claims, err := CheckToken(c)
	var valid bool
	if err != nil {
		c.JSON(401, models.Result{Code: 10001, Message: "请重新登陆"})
		valid = false
	}
	if claims.Cid == "C000000000001" {
		valid = true
	}
	if claims.Cid == cid {
		valid = true
	} else {
		c.JSON(200, models.Result{Code: 10001, Message: "您没有该操作的权限"})
		valid = false
	}
	ch <- valid
}
