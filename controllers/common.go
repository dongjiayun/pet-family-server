package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go-pet-family/config"
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

func GetAllArea(c *gin.Context) {
	type Area struct {
		Id       int    `json:"id"`
		Code     string `json:"code"`
		Name     string `json:"name"`
		Children []Area `json:"children,omitempty"`
	}
	var area []Area
	json.Unmarshal([]byte(config.AreaDict), &area)
	c.JSON(200, models.Result{0, "success", area})
}

//
//func GetPetBreedType(c *gin.Context) {
//	type petBreedGroup struct {
//		Id   int    `json:"id"`
//		Name string `json:"name"`
//	}
//	type petBreedDetail struct {
//		Id       int    `json:"id"`
//		TypeName string `json:"TypeName"`
//		IconUrl  string `json:"IconUrl"`
//	}
//	var petBreedType []PetBreedType
//	json.Unmarshal([]byte(config.PetBreedTypeDict), &petBreedType)
//	c.JSON(200, models.Result{0, "success", petBreedType})
//}
