package cmdb

import (
	"time"

	"github.com/creasty/defaults"
)

const APIVersion string = "v1alpha"

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
	Namespace         string            `json:"namespace"`
	CreateRevision    int64             `json:"create_revision"`
	Revision          int64             `json:"revision"`
	Version           int64             `json:"version"`
	ManagedFields     ManagedFields     `json:"managedFields"`
	CreationTimeStamp time.Time         `json:"creationTimestamp"`
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
	Resource `json:",inline"`
	Spec     ProjectSpec `json:"spec"`
}

type SCM struct {
	Name    string `json:"name"`
	Project string `json:"project"`
	User    string `json:"user"`
}

type AppSpec struct {
	Project string `json:"project"`
	Scm     SCM    `json:"scm"`
}

type App struct {
	Resource `json:",inline"`
	Spec     AppSpec `json:"spec"`
}
