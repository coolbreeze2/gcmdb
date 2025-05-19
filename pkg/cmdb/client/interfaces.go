package client

import "goTool/pkg/cmdb"

type Object interface {
	Read(name string, namespace string, revision int64) (map[string]any, error)
	List(opt *ListOptions) ([]map[string]any, error)
	Update(name string, namespace string, resource map[string]any) (map[string]any, error)
	Create(name string, namespace string, resource map[string]any) (map[string]any, error)
	Delete(name string, namespace string) (map[string]any, error)
	GetKind() string
	GetMetadata() cmdb.ObjectMeta
}

// TODO: Create
// TODO: Update
// TODO: Delete
