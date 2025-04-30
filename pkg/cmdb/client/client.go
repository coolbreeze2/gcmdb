package client

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
)

type Project cmdb.Project
type App cmdb.App

func NewProject() *Project {
	return &Project{
		Resource: *cmdb.NewResource("Project"),
	}
}

func (r *Project) GetKind() string {
	return r.Kind
}

func (r *Project) Read(name string, namespace string, revision int64) map[string]interface{} {
	return ReadResource(r, name, namespace, revision)
}

func (r *Project) List(opt *ListOptions) []map[string]interface{} {
	return ListResource(r, opt)
}

func (r *App) GetKind() string {
	return r.Kind
}

func NewApp() *App {
	return &App{
		Resource: *cmdb.NewResource("App"),
	}
}

func (r *App) Read(name string, namespace string, revision int64) map[string]interface{} {
	return ReadResource(r, name, namespace, revision)
}

func (r *App) List(opt *ListOptions) []map[string]interface{} {
	return ListResource(r, opt)
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

// TODO: 创建资源
// TODO: 更新资源
// TODO: 查询指定类型资源的总数 count
// TODO: 查询指定类型资源的所有名称 names

// 查询指定名称的资源详情
func ReadResource(r Object, name string, namespace string, revision int64) map[string]interface{} {
	api_url := os.Getenv("CMDB_API_URL")
	lkind := strings.ToLower(r.GetKind()) + "s"
	var url string
	if namespace == "" {
		url = path.Join(api_url, lkind, name)
	} else {
		url = path.Join(api_url, lkind, namespace, name)
	}
	query := map[string]string{"revision": strconv.FormatInt(revision, 10)}
	body := httpGet(url, query)
	var mapData map[string]interface{}
	if err := json.Unmarshal([]byte(body), &mapData); err != nil {
		panic(err)
	}
	return mapData
}

// 查询多个资源详情
func ListResource(r Object, opt *ListOptions) []map[string]interface{} {
	api_url := os.Getenv("CMDB_API_URL")
	lkind := strings.ToLower(r.GetKind()) + "s"
	var url string
	if opt.Namespace == "" {
		url = path.Join(api_url, lkind)
	} else {
		url = path.Join(api_url, lkind, opt.Namespace)
	}
	query := map[string]string{
		"page":           strconv.FormatInt(opt.Page, 10),
		"limit":          strconv.FormatInt(opt.Limit, 10),
		"selector":       EncodeSelector(opt.Selector),
		"field_selector": EncodeSelector(opt.FieldSelector),
	}
	body := httpGet(url, query)
	var mapData []map[string]interface{}
	if err := json.Unmarshal([]byte(body), &mapData); err != nil {
		panic(err)
	}
	return mapData
}

// TODO: 删除资源

func EncodeSelector(selector map[string]string) string {
	var pairs []string

	for k, v := range selector {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}

	result := strings.Join(pairs, ",")
	return result
}

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

func httpGet(url string, query map[string]string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}

	req.URL.RawQuery = q.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	statusCode := res.StatusCode
	if statusCode >= 400 {
		log.Fatalf("status code: %v\n%s", statusCode, body)
	}
	return body
}
