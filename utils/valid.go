package utils

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) ValidatePhone(fl validator.FieldLevel) bool {
	phone, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	if phone == "" {
		return true
	}
	// Regular expression to validate phone numbers (allowing only digits and dashes)
	phoneRegex := regexp.MustCompile(`^((1[3-9])[0-9]{9})|(0\d{2,3}-?\d{7,8})$`)
	return phoneRegex.MatchString(phone)
}

func (cv *CustomValidator) ValidateEmail(fl validator.FieldLevel) bool {
	email, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	if email == "" {
		return true
	}
	// Regular expression to validate email addresses
	emailRegex := regexp.MustCompile(`^\w+@[a-zA-Z0-9]{2,10}(?:\.[a-z]{2,4}){1,3}$`)
	return emailRegex.MatchString(email)
}

func CheckEmail(email string) bool {
	// Regular expression to validate email addresses
	emailRegex := regexp.MustCompile(`^\w+@[a-zA-Z0-9]{2,10}(?:\.[a-z]{2,4}){1,3}$`)
	return emailRegex.MatchString(email)
}

func CheckPhone(phone string) bool {
	// Regular expression to validate phone numbers (allowing only digits and dashes)
	phoneRegex := regexp.MustCompile(`^((1[3-9])[0-9]{9})|(0\d{2,3}-?\d{7,8})$`)
	return phoneRegex.MatchString(phone)
}

func InitValidator() {
	cv := NewCustomValidator()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("email", cv.ValidateEmail)
		v.RegisterValidation("phone", cv.ValidatePhone)
	}
}

func GetValidMsg(err error, obj interface{}) string {
	// obj为结构体指针
	getObj := reflect.TypeOf(obj)
	// 断言为具体的类型，err是一个接口
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			if f, exist := getObj.Elem().FieldByName(e.Field()); exist {
				return f.Tag.Get("msg") //错误信息不需要全部返回，当找到第一个错误的信息时，就可以结束
			}
		}
	}
	return err.Error()
}

func CheckPassword(password string) bool {
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9]+[a-zA-Z0-9!@#$%^&*()_+{}|:;\"'<>,.?/~`]{6,20}$")
	return emailRegex.MatchString(password)
}
