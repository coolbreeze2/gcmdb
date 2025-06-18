package conversion

import (
	"encoding/json"
	"fmt"
	"gcmdb/pkg/cmdb"
	"strings"

	"github.com/goccy/go-yaml"
)

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

// 解析 Selector map to string
func EncodeSelector(selector map[string]string) string {
	var pairs []string

	for k, v := range selector {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}

	result := strings.Join(pairs, ",")
	return result
}

// 解析 Selector string to map
func ParseSelector(s string) map[string]string {
	values := strings.Split(s, ",")
	_dict := map[string]string{}
	for _, value := range values {
		if value == "" {
			continue
		}
		splitedV := strings.Split(value, "=")
		_dict[splitedV[0]] = splitedV[1]
	}
	return _dict
}
