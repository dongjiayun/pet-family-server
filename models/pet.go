package models

import (
	"database/sql/driver"
	"errors"
	"github.com/goccy/go-json"
	"time"
)

type Breed struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func (breed *Breed) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*breed = Breed{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, breed)
}

func (breed Breed) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(breed)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Pet struct {
	Model
	PetId        string    `json:"petId"`
	NickName     string    `json:"nickName"`
	Age          int       `json:"age"`
	Birthday     time.Time `json:"birthday"`
	Avatar       File      `json:"avatar" gorm:"json"`
	Breed        Breed     `json:"breed" gorm:"json""`
	Gender       string    `json:"gender"`
	IsSterilized bool      `json:"isSterilized"`
}

func (pet *Pet) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*pet = Pet{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, pet)
}

func (pet Pet) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(pet)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Pets []Pet

func (pets *Pets) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*pets = Pets{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, pets)
}

func (pets Pets) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(pets)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}
