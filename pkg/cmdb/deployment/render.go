package deployment

import (
	"context"
	"encoding/json"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/conversion"
	"gcmdb/pkg/cmdb/runtime"
	"gcmdb/pkg/cmdb/server/storage"
	"maps"
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"
)

// 将 AppDeployment 引用的所有对象详情，合并至 AppDeployment 中
func resolveAppDeploymentDetail(db *storage.Store, appdeploy *cmdb.AppDeployment, appdeployDict map[string]any) (map[string]any, error) {
	var result map[string]any
	appdeployDictDp := map[string]any{}
	namespace := appdeploy.GetMeta().Namespace
	refs := runtime.GetFieldValueByTag(reflect.ValueOf(appdeploy), "", "reference")
	maps.Copy(appdeployDictDp, appdeployDict)
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
func ResolveAppDeployment(db *storage.Store, name, namespace string, params map[string]any) (*cmdb.AppDeployment, error) {
	// TODO: go-yaml 对于数组空元素存在 bug: https://github.com/goccy/go-yaml/issues/766
	var err error
	var appDeploy, resourceRange cmdb.Object
	var appdeployDict,
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
	appdeployDetailDict := map[string]any{}
	maps.Copy(appdeployDetailDict, appdeployDict)
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
	if appdeployDetailSpec, err = resolveAppDeploymentDetail(db, appDeploy.(*cmdb.AppDeployment), appdeployDetailDict); err != nil {
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

// 获取渲染后的 DeployTemplate
func ResolveDeployTemplate(db *storage.Store, name, namespace string, params map[string]any) (*cmdb.DeployTemplate, error) {
	var appDeploy *cmdb.AppDeployment
	var deployTpl cmdb.Object
	var hostNodes []cmdb.Object
	var deployTplBytes []byte
	var deployTplRedered, deployTplDeployArgs string
	var err error
	if appDeploy, err = ResolveAppDeployment(db, name, namespace, params); err != nil {
		return nil, err
	}
	if appDeploy.Spec.Template.Spec.DeployPlatform.Docker != nil {
		nodeSelector := appDeploy.Spec.Template.Spec.NodeSelector
		listOps := storage.ListOptions{LabelSelector: nodeSelector}
		if err = db.GetList(context.Background(), "HostNode", "", listOps, &hostNodes); err != nil {
			return nil, err
		}
		var hostNodesArray []map[string]any
		for _, h := range hostNodes {
			var hostNodesMap map[string]any
			if err = conversion.StructToMap(h, &hostNodesMap); err != nil {
				return nil, err
			}
			hostNodesArray = append(hostNodesArray, hostNodesMap)
		}

		params["host_nodes"] = hostNodesArray
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
	deployTplStr := string(deployTplBytes)
	// fmt.Printf("deployTplStr: %s", deployTplStr)
	deployTplRedered, err = runtime.RenderTemplate(deployTplStr, params)
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
	var objMap map[string]any
	if err = yaml.UnmarshalWithOptions([]byte(objStr), &objMap); err != nil {
		return nil, err
	}

	return mapToObject(objMap)
}

func mapToObject(objMap map[string]any) (cmdb.Object, error) {
	var err error
	var bytsObj []byte
	var obj cmdb.Object

	if bytsObj, err = json.Marshal(objMap); err != nil {
		return nil, err
	}
	if obj, err = conversion.DecodeObject(bytsObj); err != nil {
		return nil, err
	}
	return obj, nil
}
