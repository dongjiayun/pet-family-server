package main

import (
	"go-pet-family/models"
	"go-pet-family/router"
	"go-pet-family/utils"
)

func init() {
	models.InitRedis()
	utils.InitValidator()
}

func main() {
	r := router.GetRouter()
	r.Run(":8080")
}
