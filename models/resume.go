package models

import (
	"database/sql/driver"
	"errors"
	"github.com/goccy/go-json"
)

type Resume struct {
	Model
	ResumeId string `json:"resumeId" gorm:"index"`
	Content  string `json:"content" binding:"required"`
	Version  int    `json:"version" gorm:"default:1"`
	Language string `json:"language" binding:"required"`
}

type Resumes []Resume

func (resumes *Resumes) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*resumes = []Resume{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, resumes)
}

func (resumes Resumes) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(resumes)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}
