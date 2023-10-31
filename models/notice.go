package models

type Notice struct {
	Model
	NoticeId   string `json:"noticeId" gorm:"index"`
	Avatar     File   `json:"avatar" gorm:"json"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	NoticeType string `json:"noticeType" binding:"required"`
	NoticeCode string `json:"noticeCode" binding:"required"`
	TargetCid  string `json:"targetCid"`
	IsReaded   bool   `json:"isReaded"`
}

type Notices []Notice
