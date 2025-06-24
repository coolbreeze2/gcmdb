package runtime

import (
	"fmt"
	"reflect"
	"strings"
)

// 递归遍历结构体所有字段
func GetFieldValueByTag(v reflect.Value, path string, tagName string) []TagValuePair {
	refSet := map[string]bool{}
	result := []TagValuePair{}
	v = reflect.Indirect(v)
	t := v.Type()

	// 只处理结构体
	if t.Kind() != reflect.Struct {
		return result
	}

	for i := range v.NumField() {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		// 跳过不可导出字段
		if !fieldType.IsExported() {
			continue
		}

		// 构造字段的路径
		fieldTypeName := fieldType.Name
		if fieldType.Tag.Get("json") != "" {
			fieldTypeName = strings.Split(fieldType.Tag.Get("json"), ",")[0]
		}
		fieldPath := path + "." + fieldTypeName
		if path == "" {
			fieldPath = fieldTypeName
		}

		// 获取标签
		tagValue := fieldType.Tag.Get(tagName)
		if tagValue != "" {
			switch fieldValue := fieldVal.Interface().(type) {
			case string:
				if fieldValue != "" {
					result = append(result, TagValuePair{TagValue: tagValue, FieldValue: fieldValue, FieldPath: fieldPath})
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
					switch fieldValue := item.Interface().(type) {
					case string:
						if fieldValue != "" {
							result = append(result, TagValuePair{TagValue: tagValue, FieldValue: fieldValue, FieldPath: fieldPath})
						}
					}
				}
			}
		}
	}

	// 去重
	uniqResult := []TagValuePair{}
	for _, vp := range result {
		refKey := fmt.Sprintf("%s/%s", vp.TagValue, vp.FieldValue)
		if _, ok := refSet[refKey]; ok {
			continue
		}
		refSet[refKey] = true
		uniqResult = append(uniqResult, vp)
	}
	return uniqResult
}

func RecSetItem(obj map[string]any, path string, value any) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) == 1 {
		// 路径只有一级，直接赋值
		obj[parts[0]] = value
	} else {
		// 路径有多级，递归设置
		key := parts[0]
		rest := parts[1]

		// 确保对应key值是map[string]any，如果不存在则新建空map
		var nextMap map[string]any
		if v, ok := obj[key]; ok {
			nextMap, ok = v.(map[string]any)
			if !ok {
				// 如果类型不是map，覆盖为新的map
				nextMap = make(map[string]any)
				obj[key] = nextMap
			}
		} else {
			nextMap = make(map[string]any)
			obj[key] = nextMap
		}

		RecSetItem(nextMap, rest, value)
	}
}

func Merge2Dict(current, target map[string]any, path []string) map[string]any {
	if path == nil {
		path = []string{}
	}

	for key, targetVal := range target {
		if currentVal, exists := current[key]; exists {
			// 如果两个都是 map[string]any，递归合并
			currentMap, okCur := currentVal.(map[string]any)
			targetMap, okTar := targetVal.(map[string]any)
			if okCur && okTar {
				Merge2Dict(currentMap, targetMap, append(path, key))
			} else if currentVal == nil && isAllowedType(targetVal) {
				// 如果 current[key] 是 nil 并且 target[key] 是允许类型，则替换
				current[key] = targetVal
			} else if !reflect.DeepEqual(currentVal, targetVal) {
				// 值不同且 current[key] 不为空，跳过冲突处理
				// 如果需要，可以在这里抛错或打印冲突信息
				// fmt.Printf("Conflict at %s\n", strings.Join(append(path, key), "."))
			}
		} else {
			// current中不存在该key，直接赋值
			current[key] = targetVal
		}
	}
	return current
}

// 判定 target 的值是否是允许替换 current nil 的类型
func isAllowedType(v any) bool {
	switch v.(type) {
	case map[string]any, []any, string, int, int64, float32, float64:
		return true
	}
	return false
}
