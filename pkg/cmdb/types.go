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
		"secret":     NewSecret(),
		"scm":        NewSCM(),
		"datacenter": NewDatacenter(),
		"zone":       NewZone(),
		"hostnode":   NewHostNode(),
		"namespace":  NewNamespace(),
		"project":    NewProject(),
		"app":        NewApp(),
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

func NewNamespace() *Namespace {
	return &Namespace{
		ResourceBase: *NewResourceBase("Namespace", ""),
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
