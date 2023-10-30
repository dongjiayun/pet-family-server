package controllers

import (
	"github.com/gin-gonic/gin"
	"go-pet-family/models"
)

func GetObsToken(c *gin.Context) {
	type ObsTokenReq struct {
		Bucket string `json:"bucket"`
	}
	var req ObsTokenReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	ch := make(chan string)
	go models.GetObsToken(req.Bucket, ch)
	obsToken := <-ch
	c.JSON(200, models.Result{0, "success", obsToken})
}
