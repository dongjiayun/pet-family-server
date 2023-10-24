package models

type Pagination struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data"`
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
	Id         int    `json:"-" gorm:"primary_key"`
	FileId     string `json:"file_id" gorm:"index"`
	FileName   string `json:"file_name"`
	FileUrl    string `json:"file_url"`
	FileType   string `json:"file_type"`
	FileSize   int    `json:"file_size"`
	FileMd5    string `json:"file_md5"`
	CreateTime string `json:"createTime"`
	IsDeleted  bool   `json:"-"`
}

type Location struct {
	Id         int    `json:"-" gorm:"primary_key"`
	LocationId string `json:"location_id" gorm:"index"`
	Country    string `json:"country"`
	City       string `json:"city"`
	Province   string `json:"province"`
	Area       string `json:"area"`
	Street     string `json:"street"`
	StreetNum  string `json:"street_num"`
	Longitude  string `json:"longitude"`
	Latitude   string `json:"latitude"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	IsDeleted  bool   `json:"-"`
}
