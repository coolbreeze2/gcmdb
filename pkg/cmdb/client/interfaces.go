package client

type Object interface {
	Read(name string, namespace string, revision int64) map[string]interface{}
	List(opt *ListOptions) []map[string]interface{}
	GetKind() string
}
