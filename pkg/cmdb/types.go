package cmdb

import (
	"strings"
	"time"

	"github.com/creasty/defaults"
)

const APIVersion string = "v1alpha"

func NewResourceWithKind(kind string) (Resource, error) {
	kind = strings.ToLower(kind)
	kindMap := map[string]Resource{
		"secret":     NewSecret(),
		"scm":        NewSCM(),
		"datacenter": NewDatacenter(),
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

type Resource interface {
	GetKind() string
	GetMeta() ResourceMeta
}

type ManagedFields struct {
	Manager   string    `json:"manager" default:"cmctl"`
	Operation string    `json:"operation" default:"Updated"`
	Time      time.Time `json:"time"`
}

func NewManagedFields() *ManagedFields {
	obj := &ManagedFields{
		Time: time.Now(),
	}
	if err := defaults.Set(obj); err != nil {
		panic(err)
	}
	return obj
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
		ManagedFields:     *NewManagedFields(),
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
	Name    string `json:"name" validate:"required"`
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
