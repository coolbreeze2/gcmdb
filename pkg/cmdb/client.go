package cmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (r *Project) Read(name string, namespace string, revision int64) map[string]interface{} {
	return ReadResource(r, name, namespace, revision)
}

func (r *Project) List(opt *ListOptions) []map[string]interface{} {
	return ListResource(r, opt)
}

func (r *App) Read(name string, namespace string, revision int64) map[string]interface{} {
	return ReadResource(r, name, namespace, revision)
}

func (r *App) List(opt *ListOptions) []map[string]interface{} {
	return ListResource(r, opt)
}

// TODO: 创建资源
// TODO: 更新资源
// TODO: 查询指定类型资源的总数 count
// TODO: 查询指定类型资源的所有名称 names

// 查询指定名称的资源详情
func ReadResource(r IResource, name string, namespace string, revision int64) map[string]interface{} {
	api_url := os.Getenv("CMDB_API_URL")
	url := fmt.Sprintf("%s/%ss/%s", api_url, strings.ToLower(r.GetKind()), name)
	query := map[string]string{"revision": strconv.FormatInt(revision, 10)}
	body := httpGet(url, query)
	var mapData map[string]interface{}
	if err := json.Unmarshal([]byte(body), &mapData); err != nil {
		panic(err)
	}
	return mapData
}

// 查询多个资源详情
// TODO: 处理 Namespace 隔离的资源
func ListResource(r IResource, opt *ListOptions) []map[string]interface{} {
	api_url := os.Getenv("CMDB_API_URL")
	var url string
	lkind := strings.ToLower(r.GetKind())
	if opt.Namespace == "" {
		url = fmt.Sprintf("%s/%ss/%s", api_url, lkind, opt.Namespace)
	} else {
		url = fmt.Sprintf("%s/%ss", api_url, lkind)
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
