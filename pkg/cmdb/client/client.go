package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
)

func NewListOptions(
	namespace string,
	page int64,
	limit int64,
	selector map[string]string,
	field_selector map[string]string,
) *ListOptions {
	obj := &ListOptions{
		Namespace:     namespace,
		Page:          page,
		Limit:         limit,
		Selector:      selector,
		FieldSelector: field_selector,
	}
	if err := defaults.Set(obj); err != nil {
		panic(err)
	}
	return obj
}

// 创建资源
func CreateResource(r Object, name string, namespace string, resource map[string]any) (map[string]any, error) {
	var url, body string
	var statusCode int
	var err error
	var mapData map[string]any

	apiUrl := GetCMDBAPIURL()
	lkind := strings.ToLower(r.GetKind()) + "s"
	if namespace == "" {
		url, err = UrlJoin(apiUrl, lkind, "/")
	} else {
		url, err = UrlJoin(apiUrl, lkind, namespace, "/")
	}
	if err != nil {
		return nil, err
	}
	removeResourceManageFields(resource)
	body, statusCode, err = DoHttpRequest(HttpRequestArgs{Method: "POST", Url: url, Data: resource})
	if statusCode == 422 {
		return nil, ObjectValidateError{apiUrl, lkind, name, namespace, body}
	} else if statusCode == 400 {
		return nil, ObjectAlreadyExistError{apiUrl, lkind, name, namespace, body}
	} else if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(body), &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
}

// 更新资源
func UpdateResource(r Object, name string, namespace string, resource map[string]any) (map[string]any, error) {
	var url, body string
	var err error
	var mapData map[string]any

	apiUrl := GetCMDBAPIURL()
	lkind := strings.ToLower(r.GetKind()) + "s"
	if namespace == "" {
		url, err = UrlJoin(apiUrl, lkind, name)
	} else {
		url, err = UrlJoin(apiUrl, lkind, namespace, name)
	}
	if err != nil {
		return nil, err
	}
	removeResourceManageFields(resource)
	if body, _, err = DoHttpRequest(HttpRequestArgs{Method: "POST", Url: url, Data: resource}); err != nil {
		return nil, err
	}
	if body == "" {
		return mapData, nil
	}
	if err = json.Unmarshal([]byte(body), &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
}

// TODO: 查询指定类型资源的总数 count
// TODO: 查询指定类型资源的所有名称 names

// 查询指定名称的资源
func ReadResource(r Object, name string, namespace string, revision int64) (map[string]any, error) {
	var url, body string
	var statusCode int
	var err error
	var mapData map[string]any

	apiUrl := GetCMDBAPIURL()
	lkind := strings.ToLower(r.GetKind()) + "s"
	if namespace == "" {
		url, err = UrlJoin(apiUrl, lkind, name)
	} else {
		url, err = UrlJoin(apiUrl, lkind, namespace, name)
	}
	if err != nil {
		return nil, err
	}
	query := map[string]string{"revision": strconv.FormatInt(revision, 10)}
	if body, statusCode, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query}); err != nil {
		return nil, err
	} else if statusCode == 404 {
		return nil, ObjectNotFoundError{apiUrl, lkind, name, namespace}
	}
	if err = json.Unmarshal([]byte(body), &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
}

// 查询多个资源
func ListResource(r Object, opt *ListOptions) ([]map[string]any, error) {
	var url, body string
	var err error
	var mapData []map[string]any

	apiUrl := GetCMDBAPIURL()
	lkind := strings.ToLower(r.GetKind()) + "s"
	if opt.Namespace == "" {
		url, err = UrlJoin(apiUrl, lkind, "/")
	} else {
		url, err = UrlJoin(apiUrl, lkind, opt.Namespace, "/")
	}
	if err != nil {
		return nil, err
	}
	query := map[string]string{
		"page":           strconv.FormatInt(opt.Page, 10),
		"limit":          strconv.FormatInt(opt.Limit, 10),
		"selector":       EncodeSelector(opt.Selector),
		"field_selector": EncodeSelector(opt.FieldSelector),
	}
	if body, _, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query}); err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(body), &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
}

// 删除资源指定名称的资源
func DeleteResource(r Object, name string, namespace string) (map[string]any, error) {
	var url, body string
	var statusCode int
	var err error
	var mapData map[string]any

	apiUrl := GetCMDBAPIURL()
	lkind := strings.ToLower(r.GetKind()) + "s"
	if namespace == "" {
		url, err = UrlJoin(apiUrl, lkind, name)
	} else {
		url, err = UrlJoin(apiUrl, lkind, namespace, name)
	}
	if err != nil {
		return nil, err
	}

	if body, statusCode, err = DoHttpRequest(HttpRequestArgs{Method: "DELETE", Url: url}); err != nil {
		return nil, err
	} else if statusCode == 404 {
		return nil, ObjectNotFoundError{apiUrl, lkind, name, namespace}
	} else if statusCode == 204 {
		return mapData, nil
	}

	if err = json.Unmarshal([]byte(body), &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
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

// 移除系统管理字段
func removeResourceManageFields(r map[string]any) {
	fields := []string{"create_revision", "creationTimestamp", "managedFields", "revision", "version"}
	metadata := r["metadata"].(map[string]any)
	for index := range fields {
		delete(metadata, fields[index])
	}
	if metadata["namespace"] == "" {
		delete(metadata, "namespace")
	}
	r["metadata"] = metadata
}

// 发送HTTP请求
func DoHttpRequest(args HttpRequestArgs) (string, int, error) {
	// 构造URL带参数
	var request *http.Request
	var response *http.Response
	var respBody []byte
	var url_ *url.URL
	var query url.Values
	var err error

	if url_, err = url.Parse(args.Url); err != nil {
		return "", -1, err
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
			return "", -1, err
		} else {
			body = bytes.NewReader([]byte(data))
		}
	} else {
		body = bytes.NewReader(nil)
	}

	// 创建请求
	if request, err = http.NewRequest(args.Method, url_.String(), body); err != nil {
		return "", -1, err
	}

	// 添加请求头
	for k, v := range args.Headers {
		request.Header.Set(k, v)
	}

	// 使用默认客户端发起请求
	client := http.DefaultClient
	if response, err = client.Do(request); err != nil {
		return "", -1, err
	}
	defer response.Body.Close()

	// 读取响应内容
	if respBody, err = io.ReadAll(response.Body); err != nil {
		return "", -1, err
	}

	srespBody := string(respBody)
	statusCode := response.StatusCode
	if statusCode >= 400 && statusCode != 404 {
		err = ServerError{url_.String(), statusCode, srespBody}
		return srespBody, response.StatusCode, err
	}

	return srespBody, statusCode, nil
}

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

func GetCMDBAPIURL() string {
	apiUrl := os.Getenv("CMDB_API_URL")
	return apiUrl
}
