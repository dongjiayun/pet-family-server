package main

import (
	"go-pet-family/router"
	"go-pet-family/utils"
)

func main() {
	utils.InitValidator()

	r := router.GetRouter()
	r.Run(":8080")
}
