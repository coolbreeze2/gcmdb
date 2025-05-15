package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
)

type Project cmdb.Project
type App cmdb.App

var KindMap = map[string]Object{
	"Project": Project{},
	"App":     App{},
}

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

func NewProject() *Project {
	return &Project{
		Resource: *cmdb.NewResource("Project"),
	}
}

func (r Project) GetKind() string {
	return r.Kind
}

func (r Project) Read(name string, namespace string, revision int64) (map[string]any, error) {
	return ReadResource(r, name, namespace, revision)
}

func (r Project) List(opt *ListOptions) ([]map[string]any, error) {
	return ListResource(r, opt)
}

func (r Project) Update(name string, namespace string, resource map[string]any) (map[string]any, error) {
	return UpdateResource(r, name, namespace, resource)
}

func (r App) GetKind() string {
	return r.Kind
}

func NewApp() *App {
	return &App{
		Resource: *cmdb.NewResource("App"),
	}
}

func (r App) Read(name string, namespace string, revision int64) (map[string]any, error) {
	return ReadResource(r, name, namespace, revision)
}

func (r App) List(opt *ListOptions) ([]map[string]any, error) {
	return ListResource(r, opt)
}

func (r App) Update(name string, namespace string, resource map[string]any) (map[string]any, error) {
	return UpdateResource(r, name, namespace, resource)
}

// TODO: 创建资源

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
	if body, err = DoHttpRequest(HttpRequestArgs{Method: "POST", Url: url, Data: resource}); err != nil {
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
	if body, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query}); err != nil {
		return nil, err
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
	if body, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query}); err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(body), &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
}

// TODO: 删除资源

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

// 发送HTTP请求
func DoHttpRequest(args HttpRequestArgs) (string, error) {
	// 构造URL带参数
	u, err := url.Parse(args.Url)
	if err != nil {
		return "", err
	}

	// 添加查询参数
	q := u.Query()
	for k, v := range args.Query {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()

	// 创建请求体
	var body *bytes.Reader
	if args.Data != nil {
		if data, err := json.Marshal(args.Data); err != nil {
			return "", err
		} else {
			body = bytes.NewReader([]byte(data))
		}
	} else {
		body = bytes.NewReader(nil)
	}

	// 创建请求
	req, err := http.NewRequest(args.Method, u.String(), body)
	if err != nil {
		return "", err
	}

	// 添加请求头
	for k, v := range args.Headers {
		req.Header.Set(k, v)
	}

	// 使用默认客户端发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
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
