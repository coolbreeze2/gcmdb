package client

import (
	"fmt"
	"gcmdb/global"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/conversion"
	"gcmdb/pkg/cmdb/runtime"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/imroc/req/v3"
)

func (c CMDBClient) Health() bool {
	var err error
	var resp *req.Response
	url := UrlJoin(c.getCMDBAPIURL(), "/health")
	resp, err = req.C().R().Get(url)
	if err == nil && resp.StatusCode == 200 {
		return true
	}
	return false
}

// 创建资源
func (c CMDBClient) CreateResource(r cmdb.Object) (map[string]any, error) {
	var err error
	var resp *req.Response
	var resource, result map[string]any

	c.fmtError(r, resp, conversion.StructToMap(r, &resource))

	url := c.getCreateResourceUrl(r)
	RemoveResourceManageFields(resource)

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
	RemoveResourceManageFields(resource)

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
		"all":            strconv.FormatBool(opt.All),
		"page":           strconv.FormatInt(opt.Page, 10),
		"limit":          strconv.FormatInt(opt.Limit, 10),
		"selector":       conversion.EncodeSelector(opt.Selector),
		"field_selector": conversion.EncodeSelector(opt.FieldSelector),
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

// TODO: 获取渲染后的AppDeployment
func (c CMDBClient) RenderAppDeployment(name, namespace string, params map[string]any) (map[string]any, error) {
	path := fmt.Sprintf("/appdeployments/%s/%s/render", namespace, name)
	var result map[string]any
	url := c.getCMDBAPIURL() + path
	data := map[string]any{"params": map[string]any{}}
	resp, err := req.C().R().SetBody(data).SetSuccessResult(&result).SetErrorResult(&result).Post(url)
	return result, c.fmtError(&cmdb.AppDeployment{}, resp, err)
}

// TODO: 获取渲染后的 AppDeployment 关联的 DeployTemplate
func (c CMDBClient) RenderDeployTemplate(name, namespace string, params map[string]any) (map[string]any, error) {
	path := fmt.Sprintf("/appdeployments/%s/%s/deploytemplate/render", namespace, name)
	var result map[string]any
	url := c.getCMDBAPIURL() + path
	data := map[string]any{"params": map[string]any{}}
	resp, err := req.C().R().SetBody(data).SetSuccessResult(&result).SetErrorResult(&result).Post(url)
	return result, c.fmtError(&cmdb.DeployTemplate{}, resp, err)
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

// 移除系统管理字段
func RemoveResourceManageFields(r map[string]any) {
	if r == nil {
		return
	}
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

func ParseResourceFromDir(dirPath string) ([]cmdb.Object, []string, error) {
	var err error
	var objs []cmdb.Object
	var filePaths []string

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			var obj cmdb.Object
			if obj, err = ParseResourceFromFile(path); err != nil {
				return fmt.Errorf("%s\nwhen parse file %s", err.Error(), path)
			}
			objs = append(objs, obj)
			filePaths = append(filePaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return objs, filePaths, nil
}

func ParseResourceFromFile(filePath string) (cmdb.Object, error) {
	var file []byte
	var err error
	var obj cmdb.Object
	if file, err = os.ReadFile(filePath); err != nil {
		return nil, err
	}

	if obj, err = conversion.DecodeObject(file); err != nil {
		return nil, err
	}
	err = runtime.ValidateObject(obj)
	return obj, err
}
