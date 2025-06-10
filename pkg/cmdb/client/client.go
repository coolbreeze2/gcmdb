package client

import (
	"fmt"
	"goTool/global"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/conversion"
	"goTool/pkg/cmdb/runtime"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/imroc/req/v3"
)

// 创建资源
func (c CMDBClient) CreateResource(r cmdb.Object) (map[string]any, error) {
	var err error
	var resp *req.Response
	var resource, result map[string]any

	c.fmtError(r, resp, conversion.StructToMap(r, &resource))

	url := c.getCreateResourceUrl(r)
	removeResourceManageFields(resource)

	if err = runtime.ValidateObject(r); err != nil {
		return nil, err
	}

	resp, err = req.C().R().SetBody(resource).SetSuccessResult(&result).Post(url)

	return result, c.fmtError(r, resp, err)
}

// 更新资源
func (c CMDBClient) UpdateResource(r cmdb.Object) (map[string]any, error) {
	var err error
	var resp *req.Response
	var resource, result map[string]any
	meta := r.GetMeta()

	c.fmtError(r, resp, conversion.StructToMap(r, &resource))

	url := c.getURDResourceUrl(r, meta.Name, meta.Namespace)
	removeResourceManageFields(resource)

	if err = runtime.ValidateObject(r); err != nil {
		return nil, err
	}

	resp, err = req.C().R().SetBody(resource).SetSuccessResult(&result).Post(url)

	return result, c.fmtError(r, resp, err)
}

// 查询指定名称的资源
func (c CMDBClient) ReadResource(r cmdb.Object, name string, namespace string, revision int64) (map[string]any, error) {
	var err error
	var result map[string]any

	url := c.getURDResourceUrl(r, name, namespace)

	query := map[string]string{"revision": strconv.FormatInt(revision, 10)}
	resp, err := req.C().R().SetQueryParams(query).SetSuccessResult(&result).Get(url)

	return result, c.fmtError(r, resp, err)
}

// 删除资源指定名称的资源
func (c CMDBClient) DeleteResource(r cmdb.Object, name, namespace string) error {
	var err error

	url := c.getURDResourceUrl(r, name, namespace)
	resp, err := req.C().R().Delete(url)

	return c.fmtError(r, resp, err)
}

// 查询多个资源
func (c CMDBClient) ListResource(r cmdb.Object, opt *ListOptions) ([]map[string]any, error) {
	var err error
	var result []map[string]any

	url := c.getListResourceUrl(r, opt.Namespace)

	query := map[string]string{
		"page":           strconv.FormatInt(opt.Page, 10),
		"limit":          strconv.FormatInt(opt.Limit, 10),
		"selector":       EncodeSelector(opt.Selector),
		"field_selector": EncodeSelector(opt.FieldSelector),
	}
	resp, err := req.C().R().SetQueryParams(query).SetSuccessResult(&result).Get(url)

	return result, c.fmtError(r, resp, err)
}

// 查询指定类型资源的总数 count
func (c CMDBClient) CountResource(r cmdb.Object, namespace string) (int, error) {
	var err error
	var count int

	url := c.getCountResourceUrl(r)
	query := map[string]string{"namespace": namespace}
	resp, err := req.C().R().SetQueryParams(query).SetSuccessResult(&count).Get(url)

	return count, c.fmtError(r, resp, err)
}

// 查询指定类型资源的所有名称 names
func (c CMDBClient) GetResourceNames(r cmdb.Object, namespace string) ([]string, error) {
	var err error
	var names []string

	url := c.getResourceNamesUrl(r)
	query := map[string]string{"namespace": namespace}
	resp, err := req.C().R().SetQueryParams(query).SetSuccessResult(&names).Get(url)

	return names, c.fmtError(r, resp, err)
}

// 格式化错误信息
func (c CMDBClient) fmtError(r cmdb.Object, resp *req.Response, err error) error {
	if err != nil || resp == nil {
		return err
	}
	meta := r.GetMeta()
	name := meta.Name
	namespace := meta.Namespace
	lkind := LowerKind(r)
	uri := resp.Response.Request.URL.String()
	if resp.StatusCode >= 400 {
		switch resp.StatusCode {
		case 422:
			return cmdb.ResourceValidateError{Path: uri, Kind: lkind, Name: name, Namespace: namespace, Message: resp.String()}
		case 400:
			if ok, _ := regexp.MatchString("reference", resp.String()); ok {
				return cmdb.ResourceReferencedError{Path: uri, Kind: lkind, Name: name, Namespace: namespace, Message: resp.String()}
			}
			if ok, _ := regexp.MatchString("already exist", resp.String()); ok {
				return cmdb.ResourceAlreadyExistError{Path: uri, Kind: lkind, Name: name, Namespace: namespace, Message: resp.String()}
			}
		case 404:
			return cmdb.ResourceNotFoundError{Path: uri, Kind: lkind, Name: name, Namespace: namespace, Message: resp.String()}
		default:
			return cmdb.ServerError{Path: uri, StatusCode: resp.StatusCode, Message: resp.String()}
		}
	}
	return nil
}

// 更新/查询/删除的 URL
func (c CMDBClient) getURDResourceUrl(r cmdb.Object, name, namespace string) string {
	if namespace == "" {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), name)
	} else {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), namespace, name)
	}
}

// 创建的 URL
func (c CMDBClient) getCreateResourceUrl(r cmdb.Object) string {
	return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "/")
}

// 查询列表的 URL
func (c CMDBClient) getListResourceUrl(r cmdb.Object, namespace string) string {
	if namespace == "" {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "/")
	} else {
		return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), namespace)
	}
}

func (c CMDBClient) getCountResourceUrl(r cmdb.Object) string {
	return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "count", "/")
}

func (c CMDBClient) getResourceNamesUrl(r cmdb.Object) string {
	return UrlJoin(c.getCMDBAPIURL(), LowerKind(r), "names", "/")
}

func (c CMDBClient) getCMDBAPIURL() string {
	if c.ApiUrl != "" {
		return c.ApiUrl
	}
	return global.ClientSetting.CMDB_API_URL
}

func LowerKind(r cmdb.Object) string {
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

	r["metadata"] = metadata
	kind := r["kind"]
	if kind == "AppDeployment" {
		delete(r, "flow_run_id")
		delete(r, "status")
	}
}

func ParseResourceFromDir(dirPath string) ([]cmdb.Object, error) {
	var err error
	var objs []cmdb.Object
	var entries []os.DirEntry
	if entries, err = os.ReadDir(dirPath); err != nil {
		return nil, err
	}
	for _, e := range entries {
		filePath := path.Join(dirPath, e.Name())
		var obj cmdb.Object
		if obj, err = ParseResourceFromFile(filePath); err != nil {
			return nil, fmt.Errorf("%swhen parse file %s", err.Error(), filePath)
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func ParseResourceFromFile(filePath string) (cmdb.Object, error) {
	var file []byte
	var err error
	if file, err = os.ReadFile(filePath); err != nil {
		return nil, err
	}

	return conversion.DecodeObject(file)
}
