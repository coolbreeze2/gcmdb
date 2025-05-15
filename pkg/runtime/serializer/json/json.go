package json

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb/client"
	"reflect"
)

func StructToMap(obj interface{}) (map[string]interface{}, error) {
	// 检查 obj 是否是切片或数组类型
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		// 处理切片/数组类型
		var sliceMap []map[string]interface{}
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			// 递归处理每个元素
			m, err := StructToMap(elem)
			if err != nil {
				return nil, fmt.Errorf("slice/array element error: %v", err)
			}
			sliceMap = append(sliceMap, m)
		}
		// 将切片包装成 map 返回
		result := map[string]any{}
		result["items"] = sliceMap
		return result, nil
	}

	// 处理单个 struct
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func MapToStruct(data map[string]any, out client.Object, typ reflect.Type) error {
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("out must be a pointer")
	}

	if v.IsNil() {
		return fmt.Errorf("out must be a non-nil")
	}

	val := reflect.ValueOf(data)
	valKind := val.Kind()

	// 确保 data 是一个 map
	if valKind == reflect.Map {
		// 遍历传入的 map 并将值设置到结构体中
		for _, key := range val.MapKeys() {
			keyStr := key.String()
			fmt.Printf("正在设置字段%v\n", keyStr)
			// TODO: fix 这里传入的应该是 struct 而不是指针
			field, ok := FieldByTag(&out, "json", keyStr)
			if ok && field.IsValid() && field.CanSet() {
				field.Set(val.MapIndex(key).Convert(field.Type()))
			} else {
				fmt.Printf("字段%v无效或不可设置\n", keyStr)
			}
		}
	}
	return nil
}

// 根据结构体的 tag 获取 Field
func FieldByTag(v interface{}, tagKey, tagValue string) (reflect.Value, bool) {
	val := reflect.ValueOf(v)
	valKind := val.Kind()

	// 确保传入的是结构体
	if valKind != reflect.Ptr || val.IsNil() {
		return reflect.Value{}, false
	}

	// 获取指向的实际结构体类型
	v = val.Elem()
	typ := val.Type()

	// 遍历所有字段
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldKind := field.Type.Kind()

		// 获取字段标签
		tagValue_ := field.Tag.Get(tagKey) // 可以替换为需要的标签名称

		fmt.Printf("Field Name: %s, Type: %s, fieldKind: %s\n", field.Name, field.Type, fieldKind)
		// 检查标签是否匹配
		if tagValue_ == tagValue {
			return val.Field(i), true
		}

		// 判断是否是嵌入结构体
		if fieldKind == reflect.Struct {
			// 递归调用 FieldByTag 来查找嵌入结构体字段
			// TODO: 修复递归调用传递的字段（获取真实值）
			realField := val.Field(i).Interface()
			if value, found := FieldByTag(realField, tagKey, tagValue); found {
				return value, true
			}
		}
	}
	return reflect.Value{}, false
}
