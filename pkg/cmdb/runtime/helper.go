package runtime

import (
	"fmt"
	"goTool/pkg/cmdb"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// 字段校验
func ValidateObject(r cmdb.Object) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(r)
}

// 递归遍历结构体所有字段
func GetFieldValueByTag(v reflect.Value, path string, tagName string) []TagValuePair {
	result := []TagValuePair{}
	v = reflect.Indirect(v)
	t := v.Type()

	// 只处理结构体
	if t.Kind() != reflect.Struct {
		return result
	}

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		// 跳过不可导出字段
		if !fieldType.IsExported() {
			continue
		}

		// 构造字段的路径
		fieldPath := path + "." + fieldType.Name
		if path == "" {
			fieldPath = fieldType.Name
		}

		// 获取标签
		tagValue := fieldType.Tag.Get(tagName)
		if tagValue != "" {
			switch fieldValue := fieldVal.Interface().(type) {
			case string:
				if fieldValue != "" {
					result = append(result, TagValuePair{TagValue: tagValue, FieldValue: fieldValue})
				}
			}
		}

		// 递归处理嵌套结构体或数组、切片
		switch fieldVal.Kind() {
		case reflect.Struct:
			result = append(result, GetFieldValueByTag(fieldVal, fieldPath, tagName)...)
		case reflect.Ptr:
			if !fieldVal.IsNil() {
				result = append(result, GetFieldValueByTag(fieldVal.Elem(), fieldPath, tagName)...)
			}
		case reflect.Slice, reflect.Array:
			for j := 0; j < fieldVal.Len(); j++ {
				item := fieldVal.Index(j)
				itemPath := fmt.Sprintf("%s[%d]", fieldPath, j)
				if item.Kind() == reflect.Struct || (item.Kind() == reflect.Ptr && !item.IsNil()) {
					result = append(result, GetFieldValueByTag(item, itemPath, tagName)...)
				} else if tagValue != "" {
					fieldValue := item.Interface().(string)
					if fieldValue != "" {
						result = append(result, TagValuePair{TagValue: tagValue, FieldValue: fieldValue})
					}
				}
			}
		}
	}
	return result
}
