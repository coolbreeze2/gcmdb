package cmdb

import (
	"strings"
	"time"
)

const APIVersion string = "v1alpha"

type Resource interface {
	GetKind() string
	GetMeta() ResourceMeta
}

func NewResourceWithKind(kind string) (Resource, error) {
	kind = strings.ToLower(kind)
	kindMap := map[string]Resource{
		"secret":            NewSecret(),
		"scm":               NewSCM(),
		"datacenter":        NewDatacenter(),
		"zone":              NewZone(),
		"hostnode":          NewHostNode(),
		"helmrepository":    NewHelmRepository(),
		"containerregistry": NewContainerRegistry(),
		"configcenter":      NewConfigCenter(),
		"deployplatform":    NewDeployPlatform(),
		"namespace":         NewNamespace(),
		"deploytemplate":    NewDeployTemplate(),
		"project":           NewProject(),
		"app":               NewApp(),
		"resourcerange":     NewResourceRange(),
		"orchestration":     NewOrchestration(),
	}
	if r, ok := kindMap[kind]; ok {
		return r, nil
	}
	return nil, ResourceTypeError{Kind: kind}
}

func NewSecret() *Secret {
	return &Secret{
		ResourceBase: *NewResourceBase("Secret", ""),
	}
}

func NewDatacenter() *Datacenter {
	return &Datacenter{
		ResourceBase: *NewResourceBase("Datacenter", ""),
	}
}

func NewZone() *Zone {
	return &Zone{
		ResourceBase: *NewResourceBase("Zone", ""),
	}
}

func NewHostNode() *HostNode {
	return &HostNode{
		ResourceBase: *NewResourceBase("HostNode", ""),
	}
}

func NewHelmRepository() *HelmRepository {
	return &HelmRepository{
		ResourceBase: *NewResourceBase("HelmRepository", ""),
	}
}

func NewContainerRegistry() *ContainerRegistry {
	return &ContainerRegistry{
		ResourceBase: *NewResourceBase("ContainerRegistry", ""),
	}
}

func NewConfigCenter() *ConfigCneter {
	return &ConfigCneter{
		ResourceBase: *NewResourceBase("ConfigCenter", ""),
	}
}

func NewDeployPlatform() *DeployPlatform {
	return &DeployPlatform{
		ResourceBase: *NewResourceBase("DeployPlatform", ""),
	}
}

func NewNamespace() *Namespace {
	return &Namespace{
		ResourceBase: *NewResourceBase("Namespace", ""),
	}
}

func NewDeployTemplate() *DeployTemplate {
	return &DeployTemplate{
		ResourceBase: *NewResourceBase("DeployTemplate", ""),
	}
}

func NewSCM() *SCM {
	return &SCM{
		ResourceBase: *NewResourceBase("SCM", ""),
	}
}

func NewProject() *Project {
	return &Project{
		ResourceBase: *NewResourceBase("Project", ""),
	}
}

func NewApp() *App {
	return &App{
		ResourceBase: *NewResourceBase("App", ""),
	}
}

func NewResourceRange() *ResourceRange {
	return &ResourceRange{
		ResourceBase: *NewResourceBase("ResourceRange", ""),
	}
}

func NewOrchestration() *Orchestration {
	return &Orchestration{
		ResourceBase: *NewResourceBase("Orchestration", ""),
	}
}

type ManagedFields struct {
	Manager   string    `json:"manager" default:"cmctl"`
	Operation string    `json:"operation" default:"Updated"`
	Time      time.Time `json:"time"`
}

type ResourceMeta struct {
	Name              string            `json:"name" validate:"required,dns_rfc1035_label"`
	Namespace         string            `json:"namespace" validate:"omitempty,dns_rfc1035_label"`
	CreateRevision    int64             `json:"create_revision"`
	Revision          int64             `json:"revision"`
	Version           int64             `json:"version"`
	ManagedFields     ManagedFields     `json:"managedFields"`
	CreationTimeStamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
}

func NewResourceMeta(namespace string) *ResourceMeta {
	return &ResourceMeta{
		Namespace:         namespace,
		ManagedFields:     ManagedFields{},
		CreationTimeStamp: time.Now(),
		Labels:            make(map[string]string),
		Annotations:       make(map[string]string),
	}
}

type ResourceBase struct {
	APIVersion  string       `json:"apiVersion"`
	Kind        string       `json:"kind" validate:"required"`
	Metadata    ResourceMeta `json:"metadata"`
	Description string       `json:"description"`
}

func NewResourceBase(kind, namespace string) *ResourceBase {
	return &ResourceBase{
		APIVersion: APIVersion,
		Kind:       kind,
		Metadata:   *NewResourceMeta(namespace),
	}
}

type Secret struct {
	ResourceBase `json:",inline"`
	Data         map[string]string `json:"data" validate:"required"`
}

func (r Secret) GetKind() string {
	return r.Kind
}

func (r Secret) GetMeta() ResourceMeta {
	return r.Metadata
}

type DatacenterSpec struct {
	Provider   string `json:"provider" validate:"required"`
	PrivateKey string `json:"privateKey" validate:"required,dns_rfc1035_label"`
}

type Datacenter struct {
	ResourceBase `json:",inline"`
	Spec         DatacenterSpec `json:"spec" validate:"required"`
}

func (r Datacenter) GetKind() string {
	return r.Kind
}

func (r Datacenter) GetMeta() ResourceMeta {
	return r.Metadata
}

type ZoneSpec struct {
	Provider   string `json:"provider" validate:"required"`
	Datacenter string `json:"datacenter" validate:"required,dns_rfc1035_label"`
}

type Zone struct {
	ResourceBase `json:",inline"`
	Spec         ZoneSpec `json:"spec" validate:"required"`
}

func (r Zone) GetKind() string {
	return r.Kind
}

func (r Zone) GetMeta() ResourceMeta {
	return r.Metadata
}

type NamespaceSpec struct {
	BizEnv     string `json:"bizEnv" validate:"required"`
	BizUnit    string `json:"bizUnit" validate:"required"`
	Datacenter string `json:"datacenter" validate:"required,dns_rfc1035_label"`
}

type Namespace struct {
	ResourceBase `json:",inline"`
	Spec         NamespaceSpec `json:"spec" validate:"required"`
}

func (r Namespace) GetKind() string {
	return r.Kind
}

func (r Namespace) GetMeta() ResourceMeta {
	return r.Metadata
}

type ScmSpec struct {
	Datacenter string `json:"datacenter" validate:"required,dns_rfc1035_label"`
	Url        string `json:"url" validate:"required,url"`
	Service    string `json:"service" validate:"required"`
}

type SCM struct {
	ResourceBase `json:",inline"`
	Spec         ScmSpec `json:"spec" validate:"required"`
}

func (r SCM) GetKind() string {
	return r.Kind
}

func (r SCM) GetMeta() ResourceMeta {
	return r.Metadata
}

type HostNodeDisk struct {
	Name        string `json:"name" validate:"required"`
	Status      string `json:"status" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Performance string `json:"performance" validate:"required"`
	Size        int    `json:"size" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Device      string `json:"device" validate:"required"`
}

type HostNodeSpecConfiguration struct {
	Type   string         `json:"type" validate:"required"`
	Cpu    int            `json:"cpu" validate:"required"`
	Memory string         `json:"memory" validate:"required"`
	Disk   []HostNodeDisk `json:"disk" validate:"required"`
}

type HostNodeSpec struct {
	Datacenter    string                    `json:"datacenter" validate:"required,dns_rfc1035_label"`
	Zone          string                    `json:"zone" validate:"required,dns_rfc1035_label"`
	Ip            string                    `json:"ip" validate:"required,ip"`
	PublicIp      string                    `json:"publicip" validate:"required,ip"`
	Hostname      string                    `json:"hostname" validate:"required,hostname"`
	Id            string                    `json:"id" validate:"required"`
	System        string                    `json:"system" validate:"required"`
	Image         string                    `json:"image" validate:"required"`
	Category      string                    `json:"category" validate:"required"`
	AdminUser     string                    `json:"adminUser" validate:"required"`
	Port          int                       `json:"port" validate:"required"`
	Configuration HostNodeSpecConfiguration `json:"configuration" validate:"required"`
}

type HostNodeStatus struct {
	Phase         string `json:"phase" validate:"required"`
	ServiceStatus string `json:"servicestatus" validate:"required"`
}

type HostNode struct {
	ResourceBase `json:",inline"`
	Spec         HostNodeSpec   `json:"spec" validate:"required"`
	Status       HostNodeStatus `json:"status" validte:"required"`
}

func (r HostNode) GetKind() string {
	return r.Kind
}

func (r HostNode) GetMeta() ResourceMeta {
	return r.Metadata
}

type HelmRepositorySpec struct {
	Auth       string `json:"auth" validate:"required,base64"`
	Datacenter string `json:"datacenter" validate:"required,dns_rfc1035_label"`
	Url        string `json:"url" validate:"required,url"`
}

type HelmRepository struct {
	ResourceBase `json:",inline"`
	Spec         HelmRepositorySpec `json:"spec" validate:"required"`
}

func (r HelmRepository) GetKind() string {
	return r.Kind
}

func (r HelmRepository) GetMeta() ResourceMeta {
	return r.Metadata
}

type BasicAuth struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,base64"`
}

type ContainerRegistrySpec struct {
	Auth       BasicAuth `json:"auth" validate:"required"`
	Datacenter string    `json:"datacenter" validate:"required,dns_rfc1035_label"`
	Registry   string    `json:"registry" validate:"required,hostname"`
	Type       string    `json:"type" validate:"required"`
	Url        string    `json:"url" validate:"required,url"`
}

type ContainerRegistry struct {
	ResourceBase `json:",inline"`
	Spec         ContainerRegistrySpec `json:"spec" validate:"required"`
}

func (r ContainerRegistry) GetKind() string {
	return r.Kind
}

func (r ContainerRegistry) GetMeta() ResourceMeta {
	return r.Metadata
}

type ConfigCenterApollo struct {
	Auth       string `json:"auth" validate:"required,base64"`
	MetaServer string `json:"metaServer" validate:"required,url"`
	Server     string `json:"server" validate:"required,url"`
}

type ConfigCenterSpec struct {
	Apollo ConfigCenterApollo `json:"apollo" validate:"required"`
}

type ConfigCneter struct {
	ResourceBase `json:",inline"`
	Spec         ConfigCenterSpec `json:"spec" validate:"required"`
}

func (r ConfigCneter) GetKind() string {
	return r.Kind
}

func (r ConfigCneter) GetMeta() ResourceMeta {
	return r.Metadata
}

type DeployPlatformKubernetesCluster struct {
	CA     string `json:"ca" validate:"required,base64"`
	Name   string `json:"name" validate:"required"`
	Server string `json:"server" validate:"required,url"`
}

type DeployPlatformKubernetesUser struct {
	ClientCert string `json:"client_cert" validate:"required,base64"`
	ClientKey  string `json:"client_key" validate:"required,base64"`
	Name       string `json:"name" validate:"required"`
}

type DeployPlatformKubernetes struct {
	Cluster DeployPlatformKubernetesCluster `json:"cluster" validate:"required"`
	User    DeployPlatformKubernetesUser    `json:"user" validate:"required"`
}

type DeployPlatformSpec struct {
	Datacenter string                   `json:"datacenter" validate:"required,dns_rfc1035_label"`
	Kubernetes DeployPlatformKubernetes `json:"kubernetes" validate:"required"`
}

type DeployPlatform struct {
	ResourceBase `json:",inline"`
	Spec         DeployPlatformSpec `json:"spec" validate:"required"`
}

func (r DeployPlatform) GetKind() string {
	return r.Kind
}

func (r DeployPlatform) GetMeta() ResourceMeta {
	return r.Metadata
}

type DeployTemplateSpec struct {
	Command    []string `json:"command" validate:"required"`
	DeployArgs string   `json:"deployArgs" validate:"required"`
}

type DeployTemplate struct {
	ResourceBase `json:",inline"`
	Spec         DeployTemplateSpec `json:"spec" validate:"required"`
	Data         map[string]string  `json:"data" validate:"required"`
}

func (r DeployTemplate) GetKind() string {
	return r.Kind
}

func (r DeployTemplate) GetMeta() ResourceMeta {
	return r.Metadata
}

type ProjectSpec struct {
	NameInChain string `json:"nameInChain" validate:"required"`
}

type Project struct {
	ResourceBase `json:",inline"`
	Spec         ProjectSpec `json:"spec" validate:"required"`
}

func (r Project) GetKind() string {
	return r.Kind
}

func (r Project) GetMeta() ResourceMeta {
	return r.Metadata
}

type AppSCM struct {
	Name    string `json:"name" validate:"required,dns_rfc1035_label"`
	Project string `json:"project" validate:"required"`
	User    string `json:"user" validate:"required"`
}

type AppSpec struct {
	Project string `json:"project" validate:"required,dns_rfc1035_label"`
	Scm     AppSCM `json:"scm" validate:"required"`
}

type App struct {
	ResourceBase `json:",inline"`
	Spec         AppSpec `json:"spec" validate:"required"`
}

func (r App) GetKind() string {
	return r.Kind
}

func (r App) GetMeta() ResourceMeta {
	return r.Metadata
}

type AppDepConfigCenterApollo struct {
	AppId   string `json:"appId"`
	Cluster string `json:"cluster"`
	Env     string `json:"env"`
	Name    string `json:"name" validete:"omitempty,dns_rfc1035_label"`
}

type AppDepConfigCenter struct {
	Apollo AppDepConfigCenterApollo `json:"apollo"`
}

type AppDepServiceDiscoveryEureka struct {
	Application string `json:"application"`
}

type AppDepServiceDiscovery struct {
	Eureka AppDepServiceDiscoveryEureka `json:"eureka"`
}

type ApplicationDependence struct {
	ConfigCenter     AppDepConfigCenter     `json:"configCenter"`
	ServiceDiscovery AppDepServiceDiscovery `json:"serviceDiscovery"`
}

type DPHelm struct {
	Name         string `json:"name" validate:"omitempty,dns_rfc1035_label"`
	Release      string `json:"release"`
	Chart        string `json:"chart"`
	ChartVersion string `json:"chartVersion"`
}

type DPKubernetes struct {
	Name              string              `json:"name" validate:"omitempty,dns_rfc1035_label"`
	ContainerRegistry DPContainerRegistry `json:"containerRegistry"`
	Namespace         string              `json:"namespace" validate:"omitempty,dns_rfc1035_label"`
	Helm              DPHelm              `json:"helm"`
}

type DPContainerRegistry struct {
	Name    string `json:"name" validate:"omitempty,dns_rfc1035_label"`
	Project string `json:"project"`
}

type DeployPlatformDocker struct {
	ContainerRegistry DPContainerRegistry `json:"containerRegistry"`
	NodeName          string              `json:"nodeName,omitempty" validate:"omitempty,dns_rfc1035_label"`
	NodeIP            string              `json:"nodeIP,omitempty" validate:"omitempty,ip"`
	KubernetesAgent   DPKubernetes        `json:"kubernetesAgent"`
}

type ResourceRangeDeployPlatform struct {
	Docker     *DeployPlatformDocker `json:"docker,omitempty"`
	Kubernetes *DPKubernetes         `json:"kubernetes,omitempty"`
}

type ResourceRangeDeployTemplate struct {
	Name       string            `json:"name"`
	DeployArgs map[string]string `json:"deployArgs"`
	Values     map[string]any    `json:"values"`
}

type MonitoringMetrics struct {
	Path    string `json:"path"`
	Port    string `json:"port"`
	Scraped bool   `json:"scraped"`
}

type MonitoringProbeHttpGet struct {
	Path string `json:"path"`
	Port string `json:"port"`
}

type MonitoringProbe struct {
	HttpGet MonitoringProbeHttpGet `json:"httpGet"`
}

type ResourceRangeMonitoring struct {
	Metrics MonitoringMetrics `json:"metrics"`
	Probe   MonitoringProbe   `json:"probe"`
}

type ResourceLimit struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}

type ResourceRangeResources struct {
	Limit   ResourceLimit `json:"limit"`
	Request ResourceLimit `json:"request"`
}

type ServicePort struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type ResourceRangePorts struct {
	Http    ServicePort `json:"http"`
	Metrics ServicePort `json:"metrics"`
}

type ResourceRangeSpec struct {
	App                   string                      `json:"app" validate:"omitempty,dns_rfc1035_label"`
	ApplicationDependence ApplicationDependence       `json:"applicationDependence"`
	Args                  []string                    `json:"args"`
	Env                   map[string]string           `json:"env"`
	Command               []string                    `json:"command"`
	DeployPlatform        ResourceRangeDeployPlatform `json:"deployPlatform"`
	Monitoring            ResourceRangeMonitoring     `json:"monitoring"`
	NodeSelector          map[string]string           `json:"nodeSelector"`
	Project               string                      `json:"project" validate:"omitempty,dns_rfc1035_label"`
	Resources             ResourceRangeResources      `json:"resources"`
	Ports                 *ResourceRangePorts         `json:"ports,omitempty"`
}

type ResourceRange struct {
	ResourceBase   `json:",inline"`
	Spec           ResourceRangeSpec           `json:"spec" validate:"required"`
	DeployTemplate ResourceRangeDeployTemplate `json:"deployTemplate"`
}

func (r ResourceRange) GetKind() string {
	return r.Kind
}

func (r ResourceRange) GetMeta() ResourceMeta {
	return r.Metadata
}

type OrchestrationSpec struct {
	Name       string         `json:"name" validate:"required"`
	Parameters map[string]any `json:"parameters"`
}

type Orchestration struct {
	ResourceBase `json:",inline"`
	Spec         OrchestrationSpec `json:"spec" validate:"required"`
}

func (r Orchestration) GetKind() string {
	return r.Kind
}

func (r Orchestration) GetMeta() ResourceMeta {
	return r.Metadata
}
