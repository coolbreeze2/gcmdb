package client

import "goTool/pkg/cmdb"

type App cmdb.App

func NewApp() *App {
	return &App{
		Resource: *cmdb.NewResource("App"),
	}
}

func (r App) Read(name string, namespace string, revision int64) (map[string]any, error) {
	return ReadResource(r, name, namespace, revision)
}

func (r App) List(opt *ListOptions) ([]map[string]any, error) {
	return ListResource(r, opt)
}

func (r App) Update(name string, namespace string, resource map[string]any) (map[string]any, error) {
	return UpdateResource(r, name, namespace, resource)
}

func (r App) Create(name string, namespace string, resource map[string]any) (map[string]any, error) {
	return CreateResource(r, name, namespace, resource)
}

func (r App) Delete(name string, namespace string) (map[string]any, error) {
	return DeleteResource(r, name, namespace)
}

func (r App) GetKind() string {
	return r.Kind
}

func (r App) GetMetadata() cmdb.ObjectMeta {
	return r.Metadata
}
