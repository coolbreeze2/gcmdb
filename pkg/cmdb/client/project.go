package client

import "goTool/pkg/cmdb"

type Project cmdb.Project

func NewProject() *Project {
	return &Project{
		Resource: *cmdb.NewResource("Project"),
	}
}

func (r Project) GetKind() string {
	return r.Kind
}

func (r Project) GetMetadata() cmdb.ObjectMeta {
	return r.Metadata
}

func (r Project) Read(name string, namespace string, revision int64) (map[string]any, error) {
	return ReadResource(r, name, namespace, revision)
}

func (r Project) List(opt *ListOptions) ([]map[string]any, error) {
	return ListResource(r, opt)
}

func (r Project) Update(name string, namespace string, resource map[string]any) (map[string]any, error) {
	return UpdateResource(r, name, namespace, resource)
}

func (r Project) Create(name string, namespace string, resource map[string]any) (map[string]any, error) {
	return CreateResource(r, name, namespace, resource)
}

func (r Project) Delete(name string, namespace string) (map[string]any, error) {
	return DeleteResource(r, name, namespace)
}

func (r Project) Count(namespace string) (int, error) {
	return CountResource(r, namespace)
}

func (r Project) GetNames(namespace string) ([]string, error) {
	return GetResourceNames(r, namespace)
}
