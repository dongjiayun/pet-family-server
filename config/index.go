package config

import "time"

const Secret = "123456"

const DataBase = "root:Djydjydjy9418.@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=Local"

//const DataBase = "root:@tcp(1.94.65.197:3306)/pet-family?charset=utf8mb4&parseTime=True&loc=Local"

const SmtpHost = "smtp.qq.com"

const SmtpPort = 465

const SmtpUser = "1009008432@qq.com"

const TokenExpireDuration = time.Hour * 24 * 30
