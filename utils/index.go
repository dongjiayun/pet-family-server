package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

func BindJson(c *gin.Context, obj any) (err error) {
	body, _ := c.GetRawData()
	contentType := c.GetHeader("Content-Type")
	switch contentType {
	case "application/json":
		err = json.Unmarshal(body, &obj)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}
	return nil
}

func ArrayIncludes(arr []string, target string) bool {
	numSet := make(map[string]bool)
	for _, num := range arr {
		numSet[num] = true
	}
	return numSet[target]
}
