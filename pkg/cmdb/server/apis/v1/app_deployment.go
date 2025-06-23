package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/conversion"
	"gcmdb/pkg/cmdb/runtime"
	"gcmdb/pkg/cmdb/server/storage"
	"maps"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/goccy/go-yaml"
)

type RenderParams struct {
	Params map[string]any `json:"params"`
}

// TODO: run appdeployment
// TODO: read appdeployment status
// TODO: read logs
// TODO: list appdeployment image tags

func addAppRenderApi(r *chi.Mux) {
	r.Post(
		fmt.Sprintf("%s/appdeployments/{namespace}/{name}/render", PathPrefix),
		renderAppDeploymentFunc(),
	)
	r.Post(
		fmt.Sprintf("%s/appdeployments/{namespace}/{name}/deploytemplate/render", PathPrefix),
		renderDeployTemplateFunc(),
	)
}

// render appdeployment
func renderAppDeploymentFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		name := chi.URLParam(r, "name")
		namespace := chi.URLParam(r, "namespace")
		var params RenderParams
		if err = render.Decode(r, &params); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		var appDeploy *cmdb.AppDeployment
		if appDeploy, err = resolveAppDeployment(name, namespace, params.Params); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, appDeploy)
	}
}

// render deploytemplate
func renderDeployTemplateFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		name := chi.URLParam(r, "name")
		namespace := chi.URLParam(r, "namespace")
		var params RenderParams
		if err = render.Decode(r, &params); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		var appDeploy *cmdb.DeployTemplate
		if appDeploy, err = resolveDeployTemplate(name, namespace, params.Params); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, appDeploy)
	}
}

// 将 AppDeployment 引用的所有对象详情，合并至 AppDeployment 中
func resolveAppDeploymentDetail(appdeploy *cmdb.AppDeployment, appdeployDict map[string]any) (map[string]any, error) {
	var result map[string]any
	namespace := appdeploy.GetMeta().Namespace
	refs := runtime.GetFieldValueByTag(reflect.ValueOf(appdeploy), "", "reference")
	appdeployDictDp := runtime.DeepCopyMap(appdeployDict)
	for _, ref := range refs {
		if ref.FieldValue != "" {
			if !strings.HasSuffix(ref.FieldPath, ".name") {
				continue
			}
			kind := ref.TagValue
			name := ref.FieldValue
			refObj, err := cmdb.NewResourceWithKind(kind)
			if err != nil {
				return nil, err
			}
			refMeta := refObj.GetMeta()
			refMeta.Name = name
			if refMeta.HasNamespace() {
				refMeta.Namespace = namespace
			}
			if err = db.Get(context.Background(), kind, name, refMeta.Namespace, storage.GetOptions{}, &refObj); err != nil {
				return nil, err
			}
			// 设置值
			var objMap map[string]any
			if err = conversion.StructToMap(refObj, &objMap); err != nil {
				return nil, err
			}
			setPath := strings.TrimSuffix(ref.FieldPath, ".name") + ".spec"
			setValue := objMap["spec"]
			// 递归设置 map 的值
			runtime.RecSetItem(appdeployDict, setPath, setValue)
		}
	}
	result = runtime.Merge2Dict(
		appdeployDict["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any),
		appdeployDictDp["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any),
		nil,
	)
	return result, nil
}

// 将 resourceRange 与 AppDeployment 合并，获取渲染后的 AppDeployment
func resolveAppDeployment(name, namespace string, params map[string]any) (*cmdb.AppDeployment, error) {
	// TODO: go-yaml 对于数组空元素存在 bug: https://github.com/goccy/go-yaml/issues/766
	var err error
	var appDeploy, resourceRange cmdb.Object
	var appdeployDict,
		appdeployDetailDict,
		resourceRangeDict,
		appdeployRenderDetailDict,
		appDeployRenderedDict,
		appdeployDetailSpec map[string]any
	var rrName, appDeployRendered, appDeploySecRendered string
	var appDeployYaml []byte
	var deployTemplateSeted bool

	if err = db.Get(context.Background(), "AppDeployment", name, namespace, storage.GetOptions{}, &appDeploy); err != nil {
		return nil, err
	}
	if appDeploy, ok := appDeploy.(*cmdb.AppDeployment); ok {
		rrName = appDeploy.Spec.ResourceRange
		if appDeploy.Spec.Template.DeployTemplate != nil {
			deployTemplateSeted = true
		}
	}
	if err = db.Get(context.Background(), "ResourceRange", rrName, namespace, storage.GetOptions{}, &resourceRange); err != nil {
		return nil, err
	}
	if err = conversion.StructToMap(appDeploy, &appdeployDict); err != nil {
		return nil, err
	}
	if err = conversion.StructToMap(resourceRange, &resourceRangeDict); err != nil {
		return nil, err
	}
	appdeployDetailDict = runtime.DeepCopyMap(appdeployDict)
	// 将 ResourceRange 的 spec 合并至 AppDeployment.spec.template.spec
	runtime.RecSetItem(
		appdeployDetailDict,
		"spec.template.spec",
		runtime.Merge2Dict(
			appdeployDetailDict["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any),
			resourceRangeDict["spec"].(map[string]any),
			nil,
		),
	)
	// 将 ResourceRange 的 deployTemplate 合并至 AppDeployment.template.deployTemplate
	if deployTemplateSeted {
		runtime.RecSetItem(
			appdeployDetailDict,
			"spec.template.deployTemplate",
			runtime.Merge2Dict(
				appdeployDetailDict["spec"].(map[string]any)["template"].(map[string]any)["deployTemplate"].(map[string]any),
				resourceRangeDict["deployTemplate"].(map[string]any),
				nil,
			),
		)
	} else {
		runtime.RecSetItem(
			appdeployDetailDict,
			"spec.template.deployTemplate",
			resourceRangeDict["deployTemplate"].(map[string]any),
		)
	}
	// 将 AppDeployment 引用的所有对象详情，合并至 AppDeployment 中
	if appDeployYaml, err = yaml.MarshalWithOptions(appdeployDetailDict, yaml.AutoInt()); err != nil {
		return nil, err
	}
	if appdeployDetailSpec, err = resolveAppDeploymentDetail(appDeploy.(*cmdb.AppDeployment), appdeployDetailDict); err != nil {
		return nil, err
	}
	appdeployDetailDict["spec"].(map[string]any)["template"].(map[string]any)["spec"] = appdeployDetailSpec
	params["spec"] = appdeployDetailSpec
	params["metadata"] = appdeployDetailDict["metadata"]
	if appDeployRendered, err = runtime.RenderTemplate(string(appDeployYaml), params); err != nil {
		return nil, err
	}
	// 将 AppDeployment 引用的所有对象详情，合并至第一次渲染完成的 AppDeployment 中
	if err = yaml.UnmarshalWithOptions([]byte(appDeployRendered), &appDeployRenderedDict); err != nil {
		return nil, err
	}
	appdeployRenderDetailDict = runtime.Merge2Dict(
		appDeployRenderedDict["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any),
		appdeployDetailDict["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any),
		nil,
	)
	// 支持双重引用，即链式引用，例如在 ResourceRange 对象中`spec.env.APPNAME` 值为 `${ spec.app }`，
	// 当`deployTemplate.values.env`的值写作`${ spec.env }`时，也会正常解析
	params["spec"] = appdeployRenderDetailDict
	if appDeploySecRendered, err = runtime.RenderTemplate(appDeployRendered, params); err != nil {
		return nil, err
	}
	if appDeploy, err = stringToObject(appDeploySecRendered); err != nil {
		return nil, err
	}
	return appDeploy.(*cmdb.AppDeployment), err
}

// TODO: 获取渲染后的 DeployTemplate
func resolveDeployTemplate(name, namespace string, params map[string]any) (*cmdb.DeployTemplate, error) {
	var appDeploy *cmdb.AppDeployment
	var deployTpl cmdb.Object
	var hostNodes []cmdb.Object
	var deployTplBytes []byte
	var deployTplRedered, deployTplDeployArgs string
	var err error
	if appDeploy, err = resolveAppDeployment(name, namespace, params); err != nil {
		return nil, err
	}
	if appDeploy.Spec.Template.Spec.DeployPlatform.Docker != nil {
		nodeSelector := appDeploy.Spec.Template.Spec.NodeSelector
		listOps := storage.ListOptions{LabelSelector: nodeSelector}
		if err = db.GetList(context.Background(), "HostNode", "", listOps, &hostNodes); err != nil {
			return nil, err
		}
		params["host_nodes"] = hostNodes
	}
	deployTplName := appDeploy.Spec.Template.DeployTemplate.Name
	if err = db.Get(context.Background(), "DeployTemplate", deployTplName, namespace, storage.GetOptions{}, &deployTpl); err != nil {
		return nil, err
	}
	marshalOpts := []yaml.EncodeOption{yaml.AutoInt(), yaml.UseLiteralStyleIfMultiline(true)}
	if deployTplBytes, err = yaml.MarshalWithOptions(deployTpl, marshalOpts...); err != nil {
		return nil, err
	}
	values := appDeploy.Spec.Template.DeployTemplate.Values
	maps.Copy(params, values)
	deployTplRedered, err = runtime.RenderTemplate(string(deployTplBytes), params)
	deployArgs := map[string]any{}
	for k, v := range appDeploy.Spec.Template.DeployTemplate.DeployArgs {
		deployArgs[k] = v
	}
	deployTplDeployArgs, err = runtime.RenderTemplate(
		string(deployTpl.(*cmdb.DeployTemplate).Spec.DeployArgs), deployArgs,
	)
	if deployTpl, err = stringToObject(deployTplRedered); err != nil {
		return nil, err
	}
	deployTpl.(*cmdb.DeployTemplate).Spec.DeployArgs = deployTplDeployArgs
	return deployTpl.(*cmdb.DeployTemplate), nil
}

func stringToObject(objStr string) (cmdb.Object, error) {
	var err error
	var mapObj map[string]any
	var bytsObj []byte
	var obj cmdb.Object
	if err = yaml.UnmarshalWithOptions([]byte(objStr), &mapObj); err != nil {
		return nil, err
	}

	if bytsObj, err = json.Marshal(mapObj); err != nil {
		return nil, err
	}
	if obj, err = conversion.DecodeObject(bytsObj); err != nil {
		return nil, err
	}
	return obj, nil
}
