package client

import (
	"encoding/json"
	"fmt"
	"goTool/global"
	"regexp"
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

	if url, err = getCreateListResourceUrl(r, namespace); err != nil {
		return nil, err
	}
	removeResourceManageFields(resource)

	body, statusCode, err = DoHttpRequest(HttpRequestArgs{Method: "POST", Url: url, Data: resource})

	if err = fmtCURDError(r, name, namespace, body, statusCode, err); err != nil {
		return nil, err
	}

	if mapData, err = unMarshalStringToMap(body); err != nil {
		return nil, err
	}

	return mapData, nil
}

// 更新资源
func UpdateResource(r Object, name string, namespace string, resource map[string]any) (map[string]any, error) {
	var url, body string
	var statusCode int
	var err error
	var mapData map[string]any

	if url, err = getURDResourceUrl(r, name, namespace); err != nil {
		return nil, err
	}
	removeResourceManageFields(resource)

	body, statusCode, err = DoHttpRequest(HttpRequestArgs{Method: "POST", Url: url, Data: resource})

	if err = fmtCURDError(r, name, namespace, body, statusCode, err); err != nil {
		return nil, err
	}

	if statusCode == 204 {
		return mapData, nil
	}

	if mapData, err = unMarshalStringToMap(body); err != nil {
		return nil, err
	}
	return mapData, nil
}

// 查询指定名称的资源
func ReadResource(r Object, name string, namespace string, revision int64) (map[string]any, error) {
	var url, body string
	var statusCode int
	var err error
	var mapData map[string]any

	if url, err = getURDResourceUrl(r, name, namespace); err != nil {
		return nil, err
	}

	query := map[string]string{"revision": strconv.FormatInt(revision, 10)}
	body, statusCode, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query})

	if err = fmtCURDError(r, name, namespace, body, statusCode, err); err != nil {
		return nil, err
	}

	if mapData, err = unMarshalStringToMap(body); err != nil {
		return nil, err
	}
	return mapData, nil
}

// 删除资源指定名称的资源
func DeleteResource(r Object, name, namespace string) (map[string]any, error) {
	var url, body string
	var statusCode int
	var err error
	var mapData map[string]any

	if url, err = getURDResourceUrl(r, name, namespace); err != nil {
		return nil, err
	}

	body, statusCode, err = DoHttpRequest(HttpRequestArgs{Method: "DELETE", Url: url})

	if err = fmtCURDError(r, name, namespace, body, statusCode, err); err != nil {
		return nil, err
	}

	if statusCode == 204 {
		return mapData, nil
	}

	return mapData, nil
}

// 查询多个资源
func ListResource(r Object, opt *ListOptions) ([]map[string]any, error) {
	var url, body string
	var err error
	var mapData []map[string]any

	if url, err = getCreateListResourceUrl(r, opt.Namespace); err != nil {
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

	if mapData, err = unMarshalStringToArrayMap(body); err != nil {
		return nil, err
	}
	return mapData, nil
}

// 查询指定类型资源的总数 count
func CountResource(r Object, namespace string) (int, error) {
	var url, body string
	var err error
	var count int

	if url, err = getCountResourceUrl(r); err != nil {
		return count, err
	}
	query := map[string]string{"namespace": namespace}
	if body, _, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query}); err != nil {
		return count, err
	}
	if count, err = strconv.Atoi(body); err != nil {
		return count, err
	}
	return count, nil
}

// 查询指定类型资源的所有名称 names
func GetResourceNames(r Object, namespace string) ([]string, error) {
	var url, body string
	var err error
	var names []string

	if url, err = getResourceNamesUrl(r); err != nil {
		return names, err
	}
	query := map[string]string{"namespace": namespace}
	if body, _, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query}); err != nil {
		return names, err
	}
	if err = json.Unmarshal([]byte(body), &names); err != nil {
		return nil, err
	}
	return names, nil
}

// string to map
func unMarshalStringToMap(body string) (map[string]any, error) {
	var mapData map[string]any
	if err := json.Unmarshal([]byte(body), &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
}

// string to []map
func unMarshalStringToArrayMap(body string) ([]map[string]any, error) {
	var mapData []map[string]any
	if err := json.Unmarshal([]byte(body), &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
}

// 格式化 CRUD 的错误信息
func fmtCURDError(r Object, name, namespace, body string, statusCode int, err error) error {
	apiUrl := getCMDBAPIURL()
	lkind := LowerKind(r)
	switch statusCode {
	default:
		return err
	case 422:
		return ResourceValidateError{apiUrl, lkind, name, namespace, body}
	case 400:
		if ok, _ := regexp.MatchString("reference", body); ok {
			return ResourceReferencedError{apiUrl, lkind, name, namespace, body}
		}
		if ok, _ := regexp.MatchString("already exist", body); ok {
			return ResourceAlreadyExistError{apiUrl, lkind, name, namespace, body}
		}
		return err
	case 404:
		return ResourceNotFoundError{apiUrl, lkind, name, namespace}
	}
}

// 更新/查询/读取 的 URL
func getURDResourceUrl(r Object, name, namespace string) (string, error) {
	if namespace == "" {
		return UrlJoin(getCMDBAPIURL(), LowerKind(r), name)
	} else {
		return UrlJoin(getCMDBAPIURL(), LowerKind(r), namespace, name)
	}
}

// 创建/查询列表 的 URL
func getCreateListResourceUrl(r Object, namespace string) (string, error) {
	if namespace == "" {
		return UrlJoin(getCMDBAPIURL(), LowerKind(r), "/")
	} else {
		return UrlJoin(getCMDBAPIURL(), LowerKind(r), namespace, "/")
	}
}

func getCountResourceUrl(r Object) (string, error) {
	return UrlJoin(getCMDBAPIURL(), LowerKind(r), "count", "/")
}

func getResourceNamesUrl(r Object) (string, error) {
	return UrlJoin(getCMDBAPIURL(), LowerKind(r), "names", "/")
}

func LowerKind(r Object) string {
	return strings.ToLower(r.GetKind()) + "s"
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

func getCMDBAPIURL() string {
	return global.ClientSetting.CMDB_API_URL
}
