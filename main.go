package main

import (
	"github.com/gin-gonic/gin"
	"go-pet-family/models"
	"go-pet-family/router"
	"go-pet-family/utils"
	"net/http"
)

func init() {
	models.InitRedis()
	utils.InitValidator()
}

func main() {
	r := router.GetRouter()
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})
	r.Use(router.CORSMiddleware())
	r.Run("0.0.0.0:2000")
}
