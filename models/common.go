package models

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"go-pet-family/config"
	"math"
	"reflect"
	"time"
)

type Pagination struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

type Model struct {
	Id        uint      `json:"-" gorm:"primary_key"`
	CreatedAt time.Time `json:"-" gorm:"autoCreateTime" `
	UpdatedAt time.Time `json:"-" gorm:"autoUpdateTime" `
	DeletedAt time.Time `json:"-"`
	IsAudit   bool      `json:"-"`
	AuditBy   string    `json:"-" gorm:"type:varchar(255)"`
	AuditAt   time.Time `json:"-"`
	UpdateBy  string    `json:"-" gorm:"type:varchar(255)"`
}

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data,omitempty" `
}

type ListData struct {
	List       any   `json:"list"`
	TotalPage  int   `json:"totalPage"`
	TotalCount int64 `json:"totalCount"`
	PageNo     int   `json:"pageNo"`
	PageSize   int   `json:"pageSize"`
}

func maskPhoneNumber(phone string) string {
	// 实现你的手机号掩码逻辑
	// 这里的示例只保留前三位和后四位，其他位用*替代
	if phone == "" {
		return ""
	}
	masked := phone[:3] + "****" + phone[len(phone)-4:]
	return masked
}

type File struct {
	Model
	FileId   string `json:"fileId" gorm:"index"`
	FileName string `json:"file_Name"`
	FileUrl  string `json:"fileUrl"`
	FileType string `json:"fileType"`
	FileSize int    `json:"fileSize"`
	FileMd5  string `json:"fileMd5"`
}

func (file *File) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*file = File{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, file)
}

func (file File) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Location struct {
	Model
	LocationId string `json:"locationId" gorm:"index"`
	Country    string `json:"country"`
	City       string `json:"city"`
	Province   string `json:"province"`
	Area       string `json:"area"`
	Street     string `json:"street"`
	StreetNum  string `json:"streetNum"`
	Longitude  string `json:"longitude"`
	Latitude   string `json:"latitude"`
}

func (location *Location) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*location = Location{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, location)
}

func (location Location) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(location)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

func GetListData[T interface{}](list []T, pageNo int, pageSize int, totalCount int64) ListData {
	return ListData{
		List:       list,
		TotalPage:  int(math.Ceil(float64(totalCount) / float64(pageSize))),
		TotalCount: totalCount,
		PageNo:     pageNo,
		PageSize:   pageSize,
	}
}

func GetSystemQiniuToken(bucket string, ch chan error) {
	redisClient := RedisClient
	obsToken := redisClient.Get(context.Background(), "qiniu_token")
	obsTokenValue := obsToken.Val()
	if obsTokenValue != "" {
		ch <- nil
	} else {
		accessKey := config.ObsAK
		secretKey := config.ObsSK
		mac := qbox.NewMac(accessKey, secretKey)
		if bucket == "" {
			bucket = config.ObsBucket
		}
		putPolicy := storage.PutPolicy{
			Scope: bucket,
		}
		upToken := putPolicy.UploadToken(mac)
		redisClient.Set(context.Background(), "qiniu_token", upToken, 55*time.Minute)
		ch <- nil
	}
}

func GetObsToken(bucket string, ch chan string) {
	accessKey := config.ObsAK
	secretKey := config.ObsSK
	mac := qbox.NewMac(accessKey, secretKey)
	if bucket == "" {
		bucket = config.ObsBucket
	}
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(mac)
	ch <- upToken
}

func CommonCreate[T interface{}](t *T, ch chan string) {
	value := reflect.ValueOf(t).Elem()
	typ := reflect.TypeOf(t).Elem()
	// 确保t包含Model类型的字段
	if _, ok := typ.FieldByName("Model"); !ok {
		fmt.Println("T does not contain Model type")
		return
	}

	// 设置IsAudit, AuditBy和AuditAt字段的值
	isAuditField := value.FieldByName("IsAudit")
	if isAuditField.IsValid() && isAuditField.CanSet() {
		isAuditField.SetBool(true)
	}

	auditByField := value.FieldByName("AuditBy")
	if auditByField.IsValid() && auditByField.CanSet() {
		auditByField.SetString("C-ADMIN")
	}

	auditAtField := value.FieldByName("AuditAt")
	if auditAtField.IsValid() && auditAtField.CanSet() {
		auditAtField.Set(reflect.ValueOf(time.Now()))
	}

	ch <- "success"
}

func CommonUpdate[T interface{}](t *T, c *gin.Context, ch chan string) {
	value := reflect.ValueOf(t).Elem()
	typ := reflect.TypeOf(t).Elem()

	// 确保t包含Model类型的字段
	if _, ok := typ.FieldByName("Model"); !ok {
		fmt.Println("T does not contain Model type")
		return
	}

	// 设置IsAudit, AuditBy和AuditAt字段的值
	isAuditField := value.FieldByName("IsAudit")

	if isAuditField.IsValid() && isAuditField.CanSet() {
		isAuditField.SetBool(true)
	}

	auditByField := value.FieldByName("AuditBy")
	if auditByField.IsValid() && auditByField.CanSet() {
		auditByField.SetString("C-ADMIN")
	}

	auditAtField := value.FieldByName("AuditAt")
	if auditAtField.IsValid() && auditAtField.CanSet() {
		auditAtField.Set(reflect.ValueOf(time.Now()))
	}

	cid, _ := c.Get("cid")
	var user User
	DB.Where("cid = ?", cid.(string)).First(&user)
	updateByField := value.FieldByName("UpdateBy")
	if updateByField.IsValid() && updateByField.CanSet() {
		updateByField.SetString(user.Cid)
	}
	ch <- "success"
}

func SetIntoMysqlAndRedis[T interface{}](t *T, key string, ch chan string, expireTime time.Duration) {
	redisClient := RedisClient
	v := reflect.ValueOf(t).Elem()
	id := v.FieldByName(key).String()
	redisKey := key + id
	jsonString, _ := json.Marshal(t)
	redisClient.Set(context.Background(), redisKey, jsonString, expireTime)
	db := DB.Where(key+" = ?", id).First(t)
	var hasRecord bool
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			hasRecord = false
		}
	}
	if !hasRecord {
		DB.Create(&t)
	} else {
		DB.Model(t).Where(key+" = ?", id).Updates(&t)
	}
	ch <- "success"
}

func GetFromMysqlAndRedis[T interface{}](t *T, key string, ch chan string) {
	redisClient := RedisClient
	v := reflect.ValueOf(t).Elem()
	id := v.FieldByName(key).String()
	redisKey := key + id
	jsonString, _ := redisClient.Get(context.Background(), redisKey).Result()
	if jsonString != "" {
		json.Unmarshal([]byte(jsonString), &t)
		ch <- "success"
	} else {
		db := DB.Where(key+" = ?", id).First(t)
		if db.Error != nil {
			ch <- db.Error.Error()
		}
		ch <- "success"
	}
}
