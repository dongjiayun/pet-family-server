package models

import (
	"database/sql/driver"
	"errors"
	"github.com/goccy/go-json"
)

type Tag struct {
	Model
	TagId string `json:"tagId" gorm:"index;varchar(255)"`
	Label string `json:"label" binding:"required"`
}

type Comment struct {
	Model
	CommentId string   `json:"commentId" gorm:"index"`
	Content   string   `json:"content" binding:"required"`
	Author    SafeUser `json:"author"  gorm:"type:longtext"`
	AuthorId  string   `json:"authorId"`
	Location  Location `json:"location" gorm:"type:longtext"`
	LikeIds   Ids      `json:"likeIds" gorm:"type:longtext"`
	TargetId  string   `json:"targetId" binding:"required"`
	//TargetName    string     `json:"targetName" binding:"required"`
	Attachments        []File `json:"attachments" gorm:"type:longtext"`
	ArticleId          string `json:"articleId" binding:"required"`
	RootCommentId      string `json:"rootCommentId" gorm:"default:null"`
	ChildrenCommentIds Ids    `json:"childrenCommentIds"`
}

type Tags []Tag

func (tags *Tags) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*tags = []Tag{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, tags)
}

func (tags Tags) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Covers []File

func (covers *Covers) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*covers = []File{} // covers
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, covers)
}

func (covers Covers) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(covers)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Comments []Comment

func (comments *Comments) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*comments = []Comment{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, comments)
}

func (comments Comments) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(comments)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Articles []Article

func (articles *Articles) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*articles = []Article{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, articles)
}

func (articles Articles) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(articles)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Article struct {
	Model
	ArticleId     string   `json:"articleId" gorm:"index"`
	Title         string   `json:"title" binding:"required"`
	Content       string   `json:"content" binding:"required"`
	Author        SafeUser `json:"author" gorm:"type:json"`
	AuthorId      string   `json:"authorId"`
	Covers        Covers   `json:"covers" gorm:"type:json"`
	Tags          Tags     `json:"tags" gorm:"type:json"`
	Location      Location `json:"location" gorm:"type:json"`
	LikeCids      Ids      `json:"likeCids" gorm:"type:json"`
	LikesCount    int      `json:"likesCount"`
	CollectCids   Ids      `json:"collectCids" gorm:"type:json"`
	ColllectCount int      `json:"collectCount"`
	CommentIds    Ids      `json:"commentIds" gorm:"type:json"`
	CommentCount  int      `json:"commentCount"`
	IsPrivate     bool     `json:"isPrivate" gorm:"default:false"`
	isMarkdown    bool     `json:"isMarkdown" gorm:"default:false"`
}
