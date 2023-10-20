package main

import "go-pet-family/router"

func main() {
	r := router.GetRouter()
	r.Run(":8080")
}
