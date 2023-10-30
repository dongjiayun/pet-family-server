package models

import (
	"github.com/go-redis/redis/v8"
	"go-pet-family/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

func init() {
	dsn := config.DataBase
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	migErr := DB.AutoMigrate(
		&Article{},
		&Tag{},
		&Comment{},
		&User{},
		&UserExtendInfo{},
		&Notice{},
	)
	if migErr != nil {
		panic(migErr)
	}
}

var RedisClient *redis.Client

func InitRedis() {
	// 在init函数中初始化Redis客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis服务器地址
		Password: "",               // Redis服务器密码，如果有的话
		DB:       0,                // 使用的数据库编号，默认是0
	})
}
