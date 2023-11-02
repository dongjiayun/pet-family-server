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

func ArrayIncludes[T any](arr []T, value any, fn func(T) any) bool {
	for _, v := range arr {
		if value == fn(v) {
			return true
		}
	}
	return false
}

func ArrayFilter[T any](arr []T, fn func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range arr {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
func ArrayForeach[T any](arr *[]T, fn func(T) T) {
	for index, v := range *arr {
		(*arr)[index] = fn(v)
	}
}
