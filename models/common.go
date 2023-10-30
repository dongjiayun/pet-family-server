package models

import (
	"context"
	"database/sql/driver"
	"errors"
	"github.com/goccy/go-json"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"go-pet-family/config"
	"math"
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
	AuditBy   string    `json:"-" `
	AuditAt   time.Time `json:"-"`
	UpdateBy  string    `json:"-" `
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
