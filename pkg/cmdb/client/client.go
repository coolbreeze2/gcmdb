package client

import (
	"fmt"
	"goTool/global"
	"goTool/pkg/cmdb"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
	"github.com/imroc/req/v3"
)

// 创建资源
func (c CMDBClient) CreateResource(r cmdb.Resource) (map[string]any, error) {
	var err error
	var resource, result map[string]any
	meta := r.GetMeta()

	if err = StructToMap(r, &resource); err != nil {
		return nil, err
	}

	url := c.getCreateListResourceUrl(r, meta.Namespace)
	removeResourceManageFields(resource)

	resp, err := req.C().R().SetBody(resource).SetSuccessResult(&result).Post(url)

	return result, fmtError(r, resp, err)
}

// 更新资源
func (c CMDBClient) UpdateResource(r cmdb.Resource) (map[string]any, error) {
	var err error
	var resource, result map[string]any
	meta := r.GetMeta()

	if err = StructToMap(r, &resource); err != nil {
		return nil, err
	}

	url := c.getURDResourceUrl(r, meta.Name, meta.Namespace)
	removeResourceManageFields(resource)

	resp, err := req.C().R().SetBody(resource).SetSuccessResult(&result).Post(url)

	return result, fmtError(r, resp, err)
}

// 查询指定名称的资源
func (c CMDBClient) ReadResource(r cmdb.Resource, name string, namespace string, revision int64) (map[string]any, error) {
	var err error
	var result map[string]any

	url := c.getURDResourceUrl(r, name, namespace)

	query := map[string]string{"revision": strconv.FormatInt(revision, 10)}
	resp, err := req.C().R().SetQueryParams(query).SetSuccessResult(&result).Get(url)

	return result, fmtError(r, resp, err)
}

// 删除资源指定名称的资源
func (c CMDBClient) DeleteResource(r cmdb.Resource, name, namespace string) error {
	var err error

	url := c.getURDResourceUrl(r, name, namespace)
	resp, err := req.C().R().Delete(url)

	return fmtError(r, resp, err)
}

// 查询多个资源
func (c CMDBClient) ListResource(r cmdb.Resource, opt *ListOptions) ([]map[string]any, error) {
	var err error
	var result []map[string]any

	url := c.getCreateListResourceUrl(r, opt.Namespace)

	query := map[string]string{
		"page":           strconv.FormatInt(opt.Page, 10),
		"limit":          strconv.FormatInt(opt.Limit, 10),
		"selector":       EncodeSelector(opt.Selector),
		"field_selector": EncodeSelector(opt.FieldSelector),
	}
	resp, err := req.C().R().SetQueryParams(query).SetSuccessResult(&result).Get(url)

	return result, fmtError(r, resp, err)
}

// 查询指定类型资源的总数 count
func (c CMDBClient) CountResource(r cmdb.Resource, namespace string) (int, error) {
	var err error
	var count int

	url := c.getCountResourceUrl(r)
	query := map[string]string{"namespace": namespace}
	resp, err := req.C().R().SetQueryParams(query).SetSuccessResult(&count).Get(url)

	return count, fmtError(r, resp, err)
}

// 查询指定类型资源的所有名称 names
func (c CMDBClient) GetResourceNames(r cmdb.Resource, namespace string) ([]string, error) {
	var err error
	var names []string

	url := c.getResourceNamesUrl(r)
	query := map[string]string{"namespace": namespace}
	resp, err := req.C().R().SetQueryParams(query).SetSuccessResult(&names).Get(url)

	return names, fmtError(r, resp, err)
}

// 格式化错误信息
func fmtError(r cmdb.Resource, resp *req.Response, err error) error {
	if err != nil {
		return err
	}
	meta := r.GetMeta()
	name := meta.Name
	namespace := meta.Namespace
	lkind := LowerKind(r)
	if resp.StatusCode >= 400 {
		switch resp.StatusCode {
		default:
			return cmdb.ServerError{Path: resp.Request.URL.Host, StatusCode: resp.StatusCode, Message: resp.String()}
		case 422:
			return cmdb.ResourceValidateError{Path: resp.Request.URL.Host, Kind: lkind, Name: name, Namespace: namespace, Message: resp.String()}
		case 400:
			if ok, _ := regexp.MatchString("reference", resp.String()); ok {
				return cmdb.ResourceReferencedError{Path: resp.Request.URL.Host, Kind: lkind, Name: name, Namespace: namespace, Message: resp.String()}
			}
			if ok, _ := regexp.MatchString("already exist", resp.String()); ok {
				return cmdb.ResourceAlreadyExistError{Path: resp.Request.URL.Host, Kind: lkind, Name: name, Namespace: namespace, Message: resp.String()}
			}
		case 404:
			return cmdb.ResourceNotFoundError{Path: resp.Request.URL.Host, Kind: lkind, Name: name, Namespace: namespace}
		}
	}
	return nil
}

// 更新/查询/读取 的 URL
func (c CMDBClient) getURDResourceUrl(r cmdb.Resource, name, namespace string) string {
	if namespace == "" {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), name)
	} else {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), namespace, name)
	}
}

// 创建/查询列表 的 URL
func (c CMDBClient) getCreateListResourceUrl(r cmdb.Resource, namespace string) string {
	if r.GetMeta().Namespace == "" {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "/")
	} else {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), namespace, "/")
	}
}

func (c CMDBClient) getCountResourceUrl(r cmdb.Resource) string {
	return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "count", "/")
}

func (c CMDBClient) getResourceNamesUrl(r cmdb.Resource) string {
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
