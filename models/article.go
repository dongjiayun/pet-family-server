package models

type Tag struct {
	Id         int    `json:"-" gorm:"primary_key"`
	TagId      string `json:"tagId" gorm:"index;varchar(255)"`
	Label      string `json:"label"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	IsDeleted  bool   `json:"-"`
}

type Comment struct {
	Id          int      `json:"-" gorm:"primary_key"`
	CommentId   string   `json:"comment_id" gorm:"index"`
	Content     string   `json:"content"`
	Author      User     `json:"author" gorm:"foreignKey:Cid;type:varchar(255)"`
	CreateTime  string   `json:"createTime"`
	UpdateTime  string   `json:"updateTime"`
	Location    Location `json:"location" gorm:"foreignKey:LocationId;type:varchar(255)"`
	Likes       []User   `json:"likes" gorm:"foreignKey:Cid;type:varchar(255)"`
	Target      string   `json:"target"`
	TargetId    string   `json:"targetId"`
	Attachments []File   `json:"attachments" gorm:"foreignKey:FileId;type:varchar(255)"`
	IsDeleted   bool     `json:"-"`
}

type Article struct {
	Id         int       `json:"-" gorm:"primary_key"`
	ArticleId  string    `json:"article_id" gorm:"index"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Author     User      `json:"author" gorm:"foreignKey:Cid;type:varchar(255)"`
	Covers     []File    `json:"covers" gorm:"foreignKey:FileId;type:varchar(255)"`
	Tags       []Tag     `json:"tags" gorm:"foreignKey:TagId;type:varchar(255)"`
	CreateTime string    `json:"createTime"`
	UpdateTime string    `json:"updateTime"`
	Location   Location  `json:"location" gorm:"foreignKey:LocationId;type:varchar(255)"`
	Likes      []User    `json:"likes" gorm:"foreignKey:Cid;type:varchar(255)"`
	Collects   []User    `json:"collects" gorm:"foreignKey:Cid;type:varchar(255)"`
	Comments   []Comment `json:"comments" gorm:"foreignKey:CommentId;type:varchar(255)"`
	IsDeleted  bool      `json:"-"`
}
