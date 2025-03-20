package cmdb

import (
	"time"

	"github.com/creasty/defaults"
)

const APIVersion string = "v1alpha"

type ListOptions struct {
	Namespace     string            `json:"namespace"`
	Page          int64             `json:"page"`
	Limit         int64             `json:"limit"`
	Selector      map[string]string `json:"selector"`
	FieldSelector map[string]string `json:"field_selector"`
}

func NewListOptions(page int64, limit int64, selector map[string]string, field_selector map[string]string) *ListOptions {
	obj := &ListOptions{
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

type IResource interface {
	List(opt *ListOptions) []byte
	GetKind() string
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

type ObjectMeta struct {
	Name              string            `json:"name"`
	CreateRevision    int64             `json:"create_revision"`
	Revision          int64             `json:"revision"`
	Version           int64             `json:"version"`
	ManagedFields     ManagedFields     `json:"managedFields"`
	CreationTimeStamp time.Time         `json:"creationTimeStamp"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
}

func NewObjectMeta() *ObjectMeta {
	return &ObjectMeta{
		ManagedFields:     *NewManagedFields(),
		CreationTimeStamp: time.Now(),
		Labels:            make(map[string]string),
		Annotations:       make(map[string]string),
	}
}

type Resource struct {
	APIVersion  string     `json:"apiVersion"`
	Kind        string     `json:"kind"`
	Metadata    ObjectMeta `json:"metadata"`
	Description string     `json:"description"`
}

func (r *Resource) GetKind() string {
	return r.Kind
}

func NewResource(kind string) *Resource {
	return &Resource{
		APIVersion: APIVersion,
		Kind:       kind,
		Metadata:   *NewObjectMeta(),
	}
}

type ProjectSpec struct {
	NameInChain string `json:"nameInChain"`
}

type Project struct {
	Resource
	Spec ProjectSpec `json:"spec"`
}

func NewProject() *Project {
	return &Project{
		Resource: *NewResource("Project"),
	}
}
