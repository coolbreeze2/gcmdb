package client

import (
	"encoding/json"
	"fmt"
	"goTool/global"
	"goTool/pkg/cmdb"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
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
func (c CMDBClient) CreateResource(r cmdb.Resource) (map[string]any, error) {
	var url string
	var body []byte
	var err error
	var resource, result map[string]any
	meta := r.GetMeta()

	if err = StructToMap(r, &resource); err != nil {
		return nil, err
	}

	if url, err = c.getCreateListResourceUrl(r, meta.Namespace); err != nil {
		return nil, err
	}
	removeResourceManageFields(resource)

	body, _, err = DoHttpRequest(HttpRequestArgs{Method: "POST", Url: url, Data: resource})

	if err = fmtCURDError(r, meta.Name, meta.Namespace, err); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// 更新资源
func (c CMDBClient) UpdateResource(r cmdb.Resource) (map[string]any, error) {
	var url string
	var body []byte
	var statusCode int
	var err error
	var resource, result map[string]any
	meta := r.GetMeta()

	if err = StructToMap(r, &resource); err != nil {
		return nil, err
	}

	if url, err = c.getURDResourceUrl(r, meta.Name, meta.Namespace); err != nil {
		return nil, err
	}
	removeResourceManageFields(resource)

	body, statusCode, err = DoHttpRequest(HttpRequestArgs{Method: "POST", Url: url, Data: resource})

	if err = fmtCURDError(r, meta.Name, meta.Namespace, err); err != nil {
		return nil, err
	}

	if statusCode == 204 {
		return result, nil
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// 查询指定名称的资源
func (c CMDBClient) ReadResource(r cmdb.Resource, name string, namespace string, revision int64) (map[string]any, error) {
	var url string
	var body []byte
	var err error
	var result map[string]any

	if url, err = c.getURDResourceUrl(r, name, namespace); err != nil {
		return nil, err
	}

	query := map[string]string{"revision": strconv.FormatInt(revision, 10)}
	body, _, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query})

	if err = fmtCURDError(r, name, namespace, err); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// 删除资源指定名称的资源
func (c CMDBClient) DeleteResource(r cmdb.Resource, name, namespace string) error {
	var url string
	var err error

	if url, err = c.getURDResourceUrl(r, name, namespace); err != nil {
		return err
	}

	_, _, err = DoHttpRequest(HttpRequestArgs{Method: "DELETE", Url: url})

	return fmtCURDError(r, name, namespace, err)
}

// 查询多个资源
func (c CMDBClient) ListResource(r cmdb.Resource, opt *ListOptions) ([]map[string]any, error) {
	var url string
	var body []byte
	var err error
	var result []map[string]any

	if url, err = c.getCreateListResourceUrl(r, opt.Namespace); err != nil {
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

	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// 查询指定类型资源的总数 count
func (c CMDBClient) CountResource(r cmdb.Resource, namespace string) (int, error) {
	var url string
	var body []byte
	var err error
	var count int

	if url, err = c.getCountResourceUrl(r); err != nil {
		return count, err
	}
	query := map[string]string{"namespace": namespace}
	if body, _, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query}); err != nil {
		return count, err
	}
	if count, err = strconv.Atoi(string(body)); err != nil {
		return count, err
	}
	return count, nil
}

// 查询指定类型资源的所有名称 names
func (c CMDBClient) GetResourceNames(r cmdb.Resource, namespace string) ([]string, error) {
	var url string
	var body []byte
	var err error
	var names []string

	if url, err = c.getResourceNamesUrl(r); err != nil {
		return names, err
	}
	query := map[string]string{"namespace": namespace}
	if body, _, err = DoHttpRequest(HttpRequestArgs{Method: "GET", Url: url, Query: query}); err != nil {
		return names, err
	}
	if err = json.Unmarshal(body, &names); err != nil {
		return nil, err
	}
	return names, nil
}

// 格式化 CRUD 的错误信息
func fmtCURDError(r cmdb.Resource, name, namespace string, err error) error {
	lkind := LowerKind(r)
	switch e := err.(type) {
	case cmdb.ServerError:
		switch e.StatusCode {
		case 422:
			return cmdb.ResourceValidateError{Path: e.Path, Kind: lkind, Name: name, Namespace: namespace, Message: e.Message}
		case 400:
			if ok, _ := regexp.MatchString("reference", e.Message); ok {
				return cmdb.ResourceReferencedError{Path: e.Path, Kind: lkind, Name: name, Namespace: namespace, Message: e.Message}
			}
			if ok, _ := regexp.MatchString("already exist", e.Message); ok {
				return cmdb.ResourceAlreadyExistError{Path: e.Path, Kind: lkind, Name: name, Namespace: namespace, Message: e.Message}
			}
		case 404:
			return cmdb.ResourceNotFoundError{Path: e.Path, Kind: lkind, Name: name, Namespace: namespace}
		}
	}
	return err
}

// 更新/查询/读取 的 URL
func (c CMDBClient) getURDResourceUrl(r cmdb.Resource, name, namespace string) (string, error) {
	if namespace == "" {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), name)
	} else {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), namespace, name)
	}
}

// 创建/查询列表 的 URL
func (c CMDBClient) getCreateListResourceUrl(r cmdb.Resource, namespace string) (string, error) {
	if r.GetMeta().Namespace == "" {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "/")
	} else {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), namespace, "/")
	}
}

func (c CMDBClient) getCountResourceUrl(r cmdb.Resource) (string, error) {
	return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "count", "/")
}

func (c CMDBClient) getResourceNamesUrl(r cmdb.Resource) (string, error) {
	return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "names", "/")
}

func (c CMDBClient) getCMDBAPIURL() string {
	if c.ApiUrl != "" {
		return c.ApiUrl
	}
	return global.ClientSetting.CMDB_API_URL
}

func LowerKind(r cmdb.Resource) string {
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

func ParseResourceFromDir(dirPath string) ([]cmdb.Resource, error) {
	var objs []cmdb.Resource
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		filePath := path.Join(dirPath, e.Name())
		if obj, err := ParseResourceFromFile(filePath); err == nil {
			objs = append(objs, obj)
		} else {
			return nil, err
		}
	}
	return objs, nil
}

func ParseResourceFromFile(filePath string) (cmdb.Resource, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var jsonObj map[string]any
	if err = yaml.Unmarshal(file, &jsonObj); err != nil {
		return nil, err
	}
	kind := jsonObj["kind"].(string)

	o, err := cmdb.NewResourceWithKind(kind)

	if err != nil {
		return nil, err
	}

	// 不允许设置额外字段
	if err := yaml.UnmarshalWithOptions(file, o, yaml.DisallowUnknownField()); err != nil {
		return nil, fmt.Errorf("%swhen parse file %s", err.Error(), filePath)
	}

	if err = validate.Struct(o); err != nil {
		return nil, fmt.Errorf("%s when parse file %s", err.Error(), filePath)
	}

	return o, nil
}
