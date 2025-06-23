package runtime

import (
	"encoding/base64"
	"gcmdb/pkg/cmdb"

	"github.com/go-playground/validator/v10"
)

// 字段校验
func ValidateObject(r cmdb.Object) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("base64map", base64MapValidation)
	return validate.Struct(r)
}

// map value 必须为 base64 encoding 的 string
func base64MapValidation(fl validator.FieldLevel) bool {
	val := fl.Field().Interface().(map[string]string)
	for _, value := range val {
		if _, err := base64.StdEncoding.DecodeString(value); err != nil {
			return false
		}
	}

	return true
}
