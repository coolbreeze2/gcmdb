package client

type Object interface {
	Read(name string, namespace string, revision int64) (map[string]interface{}, error)
	List(opt *ListOptions) ([]map[string]interface{}, error)
	GetKind() string
}

// TODO: Create
// TODO: Update
// TODO: Delete
