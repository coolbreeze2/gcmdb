package client

import (
	"bytes"
	"encoding/json"
	"goTool/pkg/cmdb"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"path"
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

func StructToMap(s any, out *map[string]any) error {
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
func UrlJoin(baseURL string, paths ...string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// 拼接路径部分
	for _, p := range paths {
		base.Path = path.Join(base.Path, p)
	}

	// 确保路径以 / 结尾
	if strings.HasSuffix(paths[len(paths)-1], "/") {
		base.Path += "/"
	}

	return base.String(), nil
}

// 发送HTTP请求
func DoHttpRequest(args HttpRequestArgs) ([]byte, int, error) {
	// 构造URL带参数
	var request *http.Request
	var response *http.Response
	var respBody []byte
	var url_ *url.URL
	var query url.Values
	var err error

	if url_, err = url.Parse(args.Url); err != nil {
		return []byte{}, -1, err
	}

	// 添加查询参数
	query = url_.Query()
	for k, v := range args.Query {
		if v != "" {
			query.Set(k, v)
		}
	}
	url_.RawQuery = query.Encode()

	// 创建请求体
	var body *bytes.Reader
	if args.Data != nil {
		if data, err := json.Marshal(args.Data); err != nil {
			return []byte{}, -1, err
		} else {
			body = bytes.NewReader([]byte(data))
		}
	} else {
		body = bytes.NewReader(nil)
	}

	// 创建请求
	if request, err = http.NewRequest(args.Method, url_.String(), body); err != nil {
		return []byte{}, -1, err
	}

	// 添加请求头
	for k, v := range args.Headers {
		request.Header.Set(k, v)
	}

	// 使用默认客户端发起请求
	client := http.DefaultClient
	if response, err = client.Do(request); err != nil {
		return []byte{}, -1, err
	}
	defer response.Body.Close()

	// 读取响应内容
	if respBody, err = io.ReadAll(response.Body); err != nil {
		return []byte{}, -1, err
	}

	srespBody := string(respBody)
	statusCode := response.StatusCode
	if statusCode >= 400 {
		err = cmdb.ServerError{Path: url_.String(), StatusCode: statusCode, Message: srespBody}
		return respBody, response.StatusCode, err
	}

	return respBody, statusCode, nil
}
