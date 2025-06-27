package deployment

import (
	"context"
	"fmt"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/conversion"
	"gcmdb/pkg/cmdb/server/storage"
	"maps"
	"math/rand"
	"slices"
	"strings"
	"time"
	"unicode"
)

type DeployAction string

const (
	DeployRelease DeployAction = "release"
	DeployRestart DeployAction = "restart"
)

type DeployPlatformType string

const (
	DPKubernetes DeployPlatformType = "kubernetes"
	DPDocker     DeployPlatformType = "docker"
	DPUnknown    DeployPlatformType = "unknown"
)

// AppDeployment 部署逻辑实现
type DeployController struct {
	store *storage.Store
	// prefect Deployment name
	prefectDeploymentName string
	// AppDeployment name
	name string
	// AppDeployment Namespace
	namespace string
	// release | restart
	action          DeployAction
	params          map[string]any
	appDeploy       *cmdb.AppDeployment
	newAppInstances *[]cmdb.AppInstance
	// newAppInstanceRuns []
	flowRunId string
}

func NewDeployController(db *storage.Store, action DeployAction, name, namespace string, params map[string]any) *DeployController {
	c := &DeployController{
		store:     db,
		name:      name,
		namespace: namespace,
		params:    params,
		action:    action,
	}
	return c
}

func (c *DeployController) Run() (*cmdb.AppDeployment, error) {
	// TODO: 运行 AppDeployment 部署
	var err error
	if err := c.preCheck(); err != nil {
		return nil, err
	}
	// 解析 AppDeployment
	if c.appDeploy, err = ResolveAppDeployment(c.store, c.name, c.namespace, c.params); err != nil {
		return nil, err
	}
	if err = c.createNewAppInstances(); err != nil {
		return nil, err
	}
	if err = c.runPrefectDeployment(); err != nil {
		return nil, err
	}
	if err = c.setAppDeploymentStartStatus(); err != nil {
		return nil, err
	}
	if err = c.setAppInstanceStatus(); err != nil {
		return nil, err
	}
	if err = c.setAppInstanceRunStatus(); err != nil {
		return nil, err
	}
	return c.appDeploy, nil
}

func (c *DeployController) preCheck() error {
	// 预检查
	// 检查 AppDeployment 是否在运行中
	var appDeploy cmdb.Object
	if err := c.store.Get(context.Background(), "AppDeployment", c.name, c.namespace, storage.GetOptions{}, &appDeploy); err != nil {
		return err
	}
	if appDeploy, ok := appDeploy.(*cmdb.AppDeployment); ok {
		status := appDeploy.Status
		if status == cmdb.AppDeploymentDeploying {
			errMsg := "another operation (install/upgrade/rollback/uninstall) is in progress"
			return fmt.Errorf("%s", errMsg)
		} else if c.action == DeployRestart {
			switch status {
			case cmdb.AppDeploymentNoneDeployed, cmdb.AppDeploymentUninstalled:
				errMsg := fmt.Sprintf("appDeployment %s/%s status %s can't be %s.", c.namespace, c.name, status, c.action)
				return fmt.Errorf("%s", errMsg)
			}
		}
	}
	// TODO: 检查 Prefect Deployment 是否存在
	return nil
}

func (c *DeployController) createNewAppInstances() error {
	// 根据 AppDeployment 创建 AppInstance
	// TODO: 同时创建 AppInstanceRun
	insts, err := c.genAppInstance()
	if err != nil {
		return err
	}
	newAppInstances := []cmdb.AppInstance{}
	for _, inst := range *insts {
		var out cmdb.Object
		if err = c.store.Create(context.Background(), &inst, &out); err != nil {
			return err
		}
		if out, ok := out.(*cmdb.AppInstance); ok {
			newAppInstances = append(newAppInstances, *out)
		}
	}
	c.newAppInstances = &newAppInstances
	return nil
}

func (c *DeployController) runPrefectDeployment() error {
	// TODO: 运行 Prefect Deployment
	return nil
}

func (c *DeployController) setAppDeploymentStartStatus() error {
	// 更新 AppDeployment 发布启动时的状态
	var appDeploy cmdb.Object
	if err := c.store.Get(context.Background(), "AppDeployment", c.name, c.namespace, storage.GetOptions{}, &appDeploy); err != nil {
		return err
	}
	if appDeploy, ok := appDeploy.(*cmdb.AppDeployment); ok {
		appDeploy.Status = cmdb.AppDeploymentDeploying
		appDeploy.FlowRunId = c.flowRunId
	}
	if err := c.store.Update(context.Background(), appDeploy, nil); err != nil {
		return err
	}
	return nil
}

func (c *DeployController) setAppInstanceStatus() error {
	for _, inst := range *c.newAppInstances {
		var instInDB cmdb.Object
		name := inst.Metadata.Name
		if err := c.store.Get(context.Background(), "AppInstance", name, c.namespace, storage.GetOptions{}, &instInDB); err != nil {
			return err
		}
		if appDeploy, ok := instInDB.(*cmdb.AppInstance); ok {
			appDeploy.Status.FlowRunStatus = cmdb.FlowRunRunning
			appDeploy.FlowRunId = c.flowRunId
		}
		if err := c.store.Update(context.Background(), instInDB, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *DeployController) setAppInstanceRunStatus() error {
	// TODO:
	return nil
}

func (c *DeployController) genAppInstance() (*[]cmdb.AppInstance, error) {
	insts := &[]cmdb.AppInstance{}
	typ, err := c.platformType()
	if err != nil {
		return nil, err
	}
	switch typ {
	case DPKubernetes:
		insts, err = c.genKubenertesAppInstance()
	case DPDocker:
		insts, err = c.genDockerAppInstance()
	}
	if err != nil {
		return nil, err
	}
	return insts, nil
}

func (c *DeployController) genKubenertesAppInstance() (*[]cmdb.AppInstance, error) {
	var deployTemplateResolved *cmdb.DeployTemplate
	var insts []cmdb.AppInstance
	var spec, deployTemplate map[string]any
	var err error
	labels := c.appDeploy.Metadata.Labels
	namespace := c.namespace
	instName := c.genKubenertesInstanceName()
	labels["appDeployment"] = c.name
	maps.Copy(labels, c.appDeploy.Spec.Template.Metadata.Labels)

	if err = conversion.StructToMap(c.appDeploy.Spec.Template.Spec, &spec); err != nil {
		return nil, err
	}
	if err = conversion.StructToMap(c.appDeploy.Spec.Template.DeployTemplate, &deployTemplate); err != nil {
		return nil, err
	}

	if deployTemplateResolved, err = ResolveDeployTemplate(c.store, c.name, c.namespace, c.params); err != nil {
		return nil, err
	}
	deployTemplate["data"] = deployTemplateResolved.Data

	appInstDict := map[string]any{
		"kind": "AppInstance",
		"metadata": map[string]any{
			"name":      instName,
			"namespace": namespace,
			"labels":    labels,
		},
		"spec":           spec,
		"deployTemplate": deployTemplate,
		"status":         "pending",
	}
	obj, err := mapToObject(appInstDict)
	if err != nil {
		return nil, err
	}
	if inst, ok := obj.(*cmdb.AppInstance); ok {
		insts = append(insts, *inst)
	}
	return &insts, nil
}

func (c *DeployController) genDockerAppInstance() (*[]cmdb.AppInstance, error) {
	var appInstances []cmdb.AppInstance
	var hostNodes []cmdb.HostNode
	var objs []cmdb.Object
	var deployTemplateResolved *cmdb.DeployTemplate
	var spec, deployTemplate map[string]any
	var err error
	nodeSelector := c.appDeploy.Spec.Template.Spec.NodeSelector
	if nodeSelector == nil {
		errMsg := "spec.nodeSelector 字段不允许为空"
		return nil, fmt.Errorf("%s", errMsg)
	}
	labels := c.appDeploy.Metadata.Labels
	namespace := c.namespace
	labels["appDeployment"] = c.name
	maps.Copy(labels, c.appDeploy.Spec.Template.Metadata.Labels)

	listOpts := storage.ListOptions{LabelSelector: nodeSelector}
	if err := c.store.GetList(context.Background(), "HostNode", "", listOpts, &objs); err != nil {
		return nil, err
	}
	if len(objs) == 0 {
		errMsg := fmt.Sprintf("未找到匹配到节点(nodeSelector:%v)", nodeSelector)
		return nil, fmt.Errorf("%s", errMsg)
	}

	for _, o := range objs {
		if o, ok := o.(*cmdb.HostNode); ok {
			hostNodes = append(hostNodes, *o)
		}
	}
	c.filterRestartHostNode(&hostNodes)
	for _, hostNode := range hostNodes {
		nodeName := hostNode.Metadata.Name
		nodeIp := hostNode.Spec.Ip
		instName := c.genDockerInstanceName(nodeName)

		c.appDeploy.Spec.Template.Spec.DeployPlatform.Docker.NodeName = nodeName
		c.appDeploy.Spec.Template.Spec.DeployPlatform.Docker.NodeIP = nodeIp

		if err = conversion.StructToMap(c.appDeploy.Spec.Template.Spec, &spec); err != nil {
			return nil, err
		}
		if err = conversion.StructToMap(c.appDeploy.Spec.Template.DeployTemplate, &deployTemplate); err != nil {
			return nil, err
		}

		params := map[string]any{
			"host_node_name":    nodeName,
			"host_node_ip":      nodeIp,
			"app_instance_name": instName,
		}
		maps.Copy(params, c.params)
		if deployTemplateResolved, err = ResolveDeployTemplate(c.store, c.name, c.namespace, params); err != nil {
			return nil, err
		}
		deployTemplate["data"] = deployTemplateResolved.Data

		appInstDict := map[string]any{
			"kind": "AppInstance",
			"metadata": map[string]any{
				"name":      instName,
				"namespace": namespace,
				"labels":    labels,
			},
			"spec":           spec,
			"deployTemplate": deployTemplate,
			"status":         map[string]any{"flowRunStatus": "pending"},
		}
		obj, err := mapToObject(appInstDict)
		if err != nil {
			return nil, err
		}
		if inst, ok := obj.(*cmdb.AppInstance); ok {
			appInstances = append(appInstances, *inst)
		}
	}
	return &appInstances, nil
}

func (c *DeployController) genKubenertesInstanceName() string {
	randomStr := randomString(5)
	k8sCluster := c.appDeploy.Spec.Template.Spec.DeployPlatform.Kubernetes.Name
	k8sNamespace := c.appDeploy.Spec.Template.Spec.DeployPlatform.Kubernetes.Namespace
	app := c.appDeploy.Spec.Template.Spec.App
	name := fmt.Sprintf("%s--%s--%s--%s", app, k8sCluster, k8sNamespace, randomStr)
	return truncNameLeft63(name)
}

func (c *DeployController) genDockerInstanceName(hostNodeName string) string {
	randomStr := randomString(5)
	app := c.appDeploy.Spec.Template.Spec.App
	name := fmt.Sprintf("%s--%s--%s", app, hostNodeName, randomStr)
	return truncNameLeft63(name)
}

func (c *DeployController) filterRestartHostNode(hostNodes *[]cmdb.HostNode) {
	if specifyHostnode, ok := c.params["hostnode"]; ok {
		nodes := strings.Split(specifyHostnode.(string), ",")
		*hostNodes = slices.DeleteFunc(*hostNodes, func(e cmdb.HostNode) bool {
			for _, h := range nodes {
				if e.Metadata.Name == h {
					return true
				}
			}
			return false
		})
	}
}

func (c *DeployController) platformType() (DeployPlatformType, error) {
	if c.appDeploy.Spec.Template.Spec.DeployPlatform.Kubernetes != nil {
		return DPKubernetes, nil
	} else if c.appDeploy.Spec.Template.Spec.DeployPlatform.Docker != nil {
		return DPDocker, nil
	} else {
		errMsg := fmt.Sprintf("appDeployment %s/%s deploy platform %s no support.", c.namespace, c.name, DPUnknown)
		return DPUnknown, fmt.Errorf("%s", errMsg)
	}
}

// 生成随机字符串
func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" + "0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// trunc_name_left_63 截断名称保证长度小于等于63，并移除数字前缀
func truncNameLeft63(name string) string {
	const maxLen = 63

	if len(name) > maxLen {
		name = name[len(name)-maxLen:]
	}

	for _, r := range name {
		if unicode.IsDigit(r) || r == '-' {
			name = name[1:]
		} else {
			break
		}
	}
	return name
}
