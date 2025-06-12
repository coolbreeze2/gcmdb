package storage

type GetOptions struct {
	IgnoreNotFound  bool
	ResourceVersion string
}

type ListOptions struct {
	LabelSelector map[string]string
	FieldSelector map[string]string
	Page          int64
	Limit         int64
	All           bool
}
