package client

import (
	"encoding/json"
	"goTool/pkg/cmdb"
	"math/rand"
	"strings"
	"time"
)

// 生成随机字符串
func RandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
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
				currPath := strings.Join(keys[:i], ".")
				return cmdb.MapKeyPathError{KeyPath: currPath}
			}
		}
	}
	return nil
}

// URL 路径拼接
func UrlJoin(baseURL string, paths ...string) string {
	// 拼接路径部分
	for _, p := range paths {
		if !strings.HasPrefix(baseURL, "/") && !strings.HasPrefix(p, "/") {
			baseURL += "/"
		}
		baseURL += p
	}
	return baseURL
}
