package storage

import (
	"context"
	"fmt"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/conversion"
	"goTool/pkg/cmdb/runtime"
	"path"
	"reflect"
	"strings"
	"time"

	"encoding/json"

	"github.com/mcuadros/go-defaults"
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

func (s *store) Get(ctx context.Context, kind, name, namespace string, opts GetOptions, out *cmdb.Object) error {
	obj, err := cmdb.NewResourceWithKind(kind)
	meta := obj.GetMeta()
	meta.Name = name
	meta.Namespace = namespace
	if err != nil {
		return err
	}
	key := s.getStoragePath(obj)
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
	return decode(kv, out)
}

func (s *store) Create(ctx context.Context, obj cmdb.Object, out *cmdb.Object) error {
	if err := runtime.ValidateObject(obj); err != nil {
		return err
	}

	kind := obj.GetKind()
	meta := obj.GetMeta()
	key := s.getStoragePath(obj)

	if meta.Version != 0 || meta.Revision != 0 || meta.CreateRevision != 0 {
		return fmt.Errorf("resourceVersion should not be set on objects to be created")
	}

	meta.CreationTimeStamp = time.Now()
	meta.ManagedFields.Time = time.Now()
	defaults.SetDefaults(obj)
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	if err = s.createReferences(ctx, obj); err != nil {
		return err
	}

	txnResp, err := s.client.KV.Txn(ctx).If(
		notFound(key),
	).Then(
		clientv3.OpPut(key, string(data)),
	).Commit()
	if err != nil {
		return err
	}
	if !txnResp.Succeeded {
		return NewKeyExistsError(key, 0)
	}

	if out != nil {
		return s.Get(ctx, kind, meta.Name, meta.Namespace, GetOptions{}, out)
	}
	return nil
}

func (s *store) Delete(ctx context.Context, kind, name, namespace string) error {
	obj, err := cmdb.NewResourceWithKind(kind)
	if err != nil {
		return err
	}
	if err = s.Get(ctx, kind, name, namespace, GetOptions{}, &obj); err != nil {
		return err
	}
	if err = s.checkExistReferenced(ctx, obj); err != nil {
		return err
	}
	key := s.getStoragePath(obj)
	if _, err = s.client.KV.Delete(ctx, key); err != nil {
		return err
	}
	if err = s.deleteReferences(ctx, obj); err != nil {
		return err
	}
	return nil
}

// 获取资源存储路径
func (s *store) getStoragePath(obj cmdb.Object) string {
	meta := obj.GetMeta()
	key := path.Join(strings.ToLower(obj.GetKind()) + "s")
	if meta.HasNamespace() {
		key = path.Join(key, meta.Namespace)
	}
	key = path.Join(s.pathPrefix, key, meta.Name)
	return key
}

// 验证关联对象是否存在，并创建引用关系
func (s *store) createReferences(ctx context.Context, obj cmdb.Object) error {
	meta := obj.GetMeta()
	key := s.getStoragePath(obj)
	refSet := map[string]bool{}
	var refCmps []clientv3.Cmp
	var refOps []clientv3.Op
	refs := runtime.GetFieldValueByTag(reflect.ValueOf(obj), "", "reference")
	for _, ref := range refs {
		if ref.FieldValue != "" {
			refKey := path.Join(s.pathPrefix, "references", ref.TagValue, ref.FieldValue, obj.GetKind(), meta.Name)
			if _, ok := refSet[refKey]; ok {
				// 去重
				continue
			}
			refSet[refKey] = true
			refObj, err := cmdb.NewResourceWithKind(ref.TagValue)
			if err != nil {
				return err
			}
			refMeta := refObj.GetMeta()
			refMeta.Name = ref.FieldValue
			if refMeta.HasNamespace() {
				refMeta.Namespace = meta.Namespace
			}
			refTargetKey := s.getStoragePath(refObj)

			refCmps = append(refCmps, found(refTargetKey))
			refOps = append(refOps, clientv3.OpPut(refKey, ""))
		}
	}

	txnResp, err := s.client.KV.Txn(ctx).If(
		refCmps...,
	).Then(
		refOps...,
	).Commit()
	if err != nil {
		return err
	}
	if !txnResp.Succeeded {
		return NewInvalidObjError(key, fmt.Sprintf("reference object does not exists, %v.", refs))
	}
	return nil
}

// 验证关联对象是否存在，并创建引用关系
func (s *store) deleteReferences(ctx context.Context, obj cmdb.Object) error {
	meta := obj.GetMeta()
	key := s.getStoragePath(obj)
	refSet := map[string]bool{}
	var refCmps []clientv3.Cmp
	var refOps []clientv3.Op
	refs := runtime.GetFieldValueByTag(reflect.ValueOf(obj), "", "reference")
	for _, ref := range refs {
		if ref.FieldValue != "" {
			refKey := path.Join(s.pathPrefix, "references", ref.TagValue, ref.FieldValue, obj.GetKind(), meta.Name)
			if _, ok := refSet[refKey]; ok {
				// 去重
				continue
			}
			refSet[refKey] = true
			refObj, err := cmdb.NewResourceWithKind(ref.TagValue)
			if err != nil {
				return err
			}
			refMeta := refObj.GetMeta()
			refMeta.Name = ref.FieldValue
			if refMeta.HasNamespace() {
				refMeta.Namespace = meta.Namespace
			}
			refTargetKey := s.getStoragePath(refObj)

			refCmps = append(refCmps, found(refTargetKey))
			refOps = append(refOps, clientv3.OpDelete(refKey))
		}
	}

	txnResp, err := s.client.KV.Txn(ctx).If(
		refCmps...,
	).Then(
		refOps...,
	).Commit()
	if err != nil {
		return err
	}
	if !txnResp.Succeeded {
		return NewInvalidObjError(key, fmt.Sprintf("reference object does not exists, %v.", refs))
	}
	return nil
}

// 检查当前对象是否被引用
func (s *store) checkExistReferenced(ctx context.Context, obj cmdb.Object) error {
	kind := obj.GetKind()
	name := obj.GetMeta().Name
	key := path.Join(s.pathPrefix, "references", kind, name) + "/"
	getResp, err := s.client.KV.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return err
	}
	if len(getResp.Kvs) == 0 {
		return nil
	}
	refKey := string(getResp.Kvs[0].Key)
	splitedRefKey := strings.Split(refKey, "/")
	refKind := splitedRefKey[len(splitedRefKey)-2]
	refName := splitedRefKey[len(splitedRefKey)-1]
	errMsg := fmt.Sprintf("Resource %s %s has been referenced by %s %s", kind, name, refKind, refName)
	return NewResourceReferencedError(key, errMsg)
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

func notFound(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.ModRevision(key), "=", 0)
}

func found(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.ModRevision(key), "!=", 0)
}
