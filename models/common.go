package models

import (
	"database/sql/driver"
	"errors"
	"github.com/goccy/go-json"
	"math"
	"time"
)

type Pagination struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

type Model struct {
	Id        uint       `json:"-" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-" gorm:"autoCreateTime" `
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
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
	FileId   string `json:"file_id" gorm:"index"`
	FileName string `json:"file_name"`
	FileUrl  string `json:"file_url"`
	FileType string `json:"file_type"`
	FileSize int    `json:"file_size"`
	FileMd5  string `json:"file_md5"`
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
	LocationId string `json:"location_id" gorm:"index"`
	Country    string `json:"country"`
	City       string `json:"city"`
	Province   string `json:"province"`
	Area       string `json:"area"`
	Street     string `json:"street"`
	StreetNum  string `json:"street_num"`
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
