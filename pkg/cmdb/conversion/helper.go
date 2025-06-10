package conversion

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb"
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"
)

// EnforcePtr ensures that obj is a pointer of some sort. Returns a reflect.Value
// of the dereferenced pointer, ensuring that it is settable/addressable.
// Returns an error if this is not possible.
func EnforcePtr(obj any) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got nil")
	}
	return v.Elem(), nil
}

func DecodeObject(b []byte) (cmdb.Object, error) {
	var r cmdb.Object
	var err error
	var jsonObj map[string]any
	if err = yaml.Unmarshal(b, &jsonObj); err != nil {
		return nil, err
	}
	kind := jsonObj["kind"].(string)

	if r, err = cmdb.NewResourceWithKind(kind); err != nil {
		return nil, err
	}

	// 不允许设置额外字段
	if err := yaml.UnmarshalWithOptions(b, r, yaml.DisallowUnknownField()); err != nil {
		return nil, err
	}

	return r, nil
}

func StructToMap(s any, out any) error {
	// 先将 struct 转为 JSON
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	// 再将 JSON 解析到 map
	return json.Unmarshal(data, out)
}

// Path walks the dot-delimited `path` to return a nested map value, or nil.
func GetMapValueByPath(m map[string]any, path string) any {
	var curr any = m
	var val any = nil

	keys := strings.Split(path, ".")
	for _, key := range keys {
		if nextMap, ok := curr.(map[string]any); ok {
			curr = nextMap[key]
			val = curr
		} else {
			return nil
		}
	}

	return val
}

func SetMapValueByPath(m map[string]any, path string, value any) error {
	keys := strings.Split(path, ".")
	curr := m

	for i, key := range keys {
		if i == len(keys)-1 {
			curr[key] = value
		} else {
			if nextMap, ok := curr[key].(map[string]any); ok {
				curr = nextMap
			} else {
				return cmdb.MapKeyPathError{KeyPath: path}
			}
		}
	}
	return nil
}
