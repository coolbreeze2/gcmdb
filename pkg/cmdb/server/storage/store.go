package storage

import (
	"context"
	"fmt"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/conversion"
	"path"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type store struct {
	client     *clientv3.Client
	pathPrefix string
}

func New(c *clientv3.Client, prefix string) *store {
	return newStore(c, prefix)
}

func newStore(c *clientv3.Client, prefix string) *store {
	result := &store{
		client:     c,
		pathPrefix: prefix,
	}
	return result
}

func (s *store) Get(ctx context.Context, key string, opts GetOptions, out *cmdb.Object) error {
	key = path.Join(s.pathPrefix, key)
	getResp, err := s.client.KV.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(getResp.Kvs) == 0 {
		if opts.IgnoreNotFound {
			return nil
		}
		return NewKeyNotFoundError(key, 0)
	}
	kv := getResp.Kvs[0]
	err = decode(kv, out)
	return err
}

// decode decodes value of bytes into object. It will also set the object resource version.
// On success, objPtr would be set to the object.
func decode(keyValue *mvccpb.KeyValue, objPtr *cmdb.Object) error {
	var err error
	if _, err = conversion.EnforcePtr(objPtr); err != nil {
		return fmt.Errorf("unable to convert output object to pointer: %v", err)
	}
	obj, err := conversion.DecodeObject(keyValue.Value)
	if err != nil {
		return err
	}
	setObjectVersion(keyValue, obj)
	*objPtr = obj
	return nil
}

// set the object resource version.
func setObjectVersion(keyValue *mvccpb.KeyValue, obj cmdb.Object) {
	meta := obj.GetMeta()
	meta.Version = keyValue.Version
	meta.CreateRevision = keyValue.CreateRevision
	meta.Revision = keyValue.ModRevision
}
