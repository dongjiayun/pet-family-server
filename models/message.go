package models

type Message struct {
	Model
	Content   string   `json:"content" binding:"required"`
	MessageId string   `json:"messageId"`
	AuthorId  string   `json:"authorId"`
	Author    SafeUser `json:"author" gorm:"type:json"`
	LikeIds   Ids      `json:"likeIds"`
}
