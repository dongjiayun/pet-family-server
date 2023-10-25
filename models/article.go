package models

type Tag struct {
	Model
	TagId string `json:"tagId" gorm:"index;varchar(255)"`
	Label string `json:"label"`
}

type Comment struct {
	Model
	CommentId   string   `json:"comment_id" gorm:"index"`
	Content     string   `json:"content"`
	Author      User     `json:"author" gorm:"foreignKey:Cid;type:varchar(255)"`
	Location    Location `json:"location" gorm:"foreignKey:LocationId;type:varchar(255)"`
	Likes       []User   `json:"likes" gorm:"foreignKey:Cid;type:varchar(255)"`
	Target      string   `json:"target"`
	TargetId    string   `json:"targetId"`
	Attachments []File   `json:"attachments" gorm:"foreignKey:FileId;type:varchar(255)"`
}

type Article struct {
	Model
	ArticleId  string     `json:"article_id" gorm:"index"`
	Title      string     `json:"title" binding:"required"`
	Content    string     `json:"content" binding:"required"`
	Author     SafeUser   `json:"author" gorm:"type:varchar(255)"`
	AuthorId   string     `json:"-" gorm:"type:varchar(255)"`
	Covers     []File     `json:"covers" gorm:"type:varchar(255)"`
	CoverIds   []string   `json:"coverIds" gorm:"type:varchar(255)"`
	Tags       []Tag      `json:"tags" gorm:"type:varchar(255)"`
	TagIds     []string   `json:"tagIds" gorm:"type:varchar(255)"`
	Location   Location   `json:"location" gorm:"type:varchar(255)"`
	LocationId string     `json:"locationId" gorm:"type:varchar(255)"`
	Likes      []SafeUser `json:"likes" gorm:"type:varchar(255)"`
	LikeIds    []string   `json:"likeIds" gorm:"type:varchar(255)"`
	Collects   []SafeUser `json:"collects" gorm:"type:varchar(255)"`
	CollectIds []string   `json:"collectIds" gorm:"type:varchar(255)"`
	Comments   []Comment  `json:"comments" gorm:"type:varchar(255)"`
	CommentIds []string   `json:"commentIds" gorm:"type:varchar(255)"`
}
