package models

import "time"

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

func HasFile(fileId string) bool {
	db := DB.Where("file_id = ?", fileId).Where("is_deleted = 0").First(&File{})
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		return false
	}
	return true
}

func CreateFile(file *File) error {
	hasFile := HasFile(file.FileId)
	if hasFile {
		return nil
	}
	db := DB.Create(file)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		return db.Error
	}
	return nil
}
