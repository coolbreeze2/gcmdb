package storage

import (
	"bytes"
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

type referenceAction int

const (
	referenceActionCreate referenceAction = iota + 1
	referenceActionDelete
	referenceActionCheckExist
)

type Store struct {
	client     *clientv3.Client
	pathPrefix string
}

func New(c *clientv3.Client, prefix string) *Store {
	return newStore(c, prefix)
}

func newStore(c *clientv3.Client, prefix string) *Store {
	result := &Store{
		client:     c,
		pathPrefix: prefix,
	}
	return result
}

func (s *Store) Health(ctx context.Context) bool {
	_, err := s.client.Status(ctx, s.client.Endpoints()[0])
	return err == nil
}

func (s *Store) Get(ctx context.Context, kind, name, namespace string, opts GetOptions, out *cmdb.Object) error {
	obj, err := cmdb.NewResourceWithKind(kind)
	if err != nil {
		return err
	}
	meta := obj.GetMeta()
	meta.Name = name
	meta.Namespace = namespace
	key := s.getStoragePath(obj)
	getResp, err := s.client.KV.Get(ctx, key)
	if err != nil {
		return NewInternalError(err.Error())
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

func (s *Store) Count(ctx context.Context, kind, namespace string) (int64, error) {
	key := s.getStoragePathPrefix(kind, namespace, false)
	getResp, err := s.client.KV.Get(ctx, key, clientv3.WithRange(clientv3.GetPrefixRangeEnd(key)), clientv3.WithCountOnly())
	if err != nil {
		return 0, NewInternalError(err.Error())
	}
	return getResp.Count, nil
}

func (s *Store) GetNames(ctx context.Context, kind, namespace string) ([]string, error) {
	names := []string{}
	key := s.getStoragePathPrefix(kind, namespace, false)
	getResp, err := s.client.KV.Get(ctx, key, clientv3.WithRange(clientv3.GetPrefixRangeEnd(key)), clientv3.WithKeysOnly())
	if err != nil {
		return names, NewInternalError(err.Error())
	}
	for _, kv := range getResp.Kvs {
		splitedKey := strings.Split(string(kv.Key), "/")
		names = append(names, splitedKey[len(splitedKey)-1])
	}
	return names, nil
}

func (s *Store) GetList(ctx context.Context, kind, namespace string, opts ListOptions, out *[]cmdb.Object) error {
	key := s.getStoragePathPrefix(kind, namespace, opts.All)
	if opts.All {
		opts.Limit = 0
	}
	var minCreateRevision, maxCreateRevision int64
	if len(opts.LabelSelector) == 0 && len(opts.FieldSelector) == 0 {
		rangeLimit := opts.Page * opts.Limit
		ops := []clientv3.OpOption{
			clientv3.WithLimit(rangeLimit),
			clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
			clientv3.WithKeysOnly(),
			clientv3.WithPrefix(),
		}
		keysResp, err := s.client.KV.Get(ctx, key, ops...)
		if err != nil {
			return NewInternalError(err.Error())
		}
		count := keysResp.Count
		var minKVIndex int64
		if opts.Page > 1 {
			minKVIndex = (opts.Page - 1) * opts.Limit
		}
		maxKVIndex := minKVIndex + opts.Limit + 1
		if maxKVIndex >= count {
			maxKVIndex = count - 1
		}
		if minKVIndex+1 > count {
			return nil
		}
		minKV := keysResp.Kvs[minKVIndex]
		maxKV := keysResp.Kvs[maxKVIndex]
		minCreateRevision = minKV.CreateRevision
		maxCreateRevision = maxKV.CreateRevision
	}
	ops := []clientv3.OpOption{
		clientv3.WithMinCreateRev(minCreateRevision),
		clientv3.WithMaxCreateRev(maxCreateRevision),
		clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
		clientv3.WithPrefix(),
	}
	kvResp, err := s.client.KV.Get(ctx, key, ops...)
	if err != nil {
		return NewInternalError(err.Error())
	}
	for _, kvs := range kvResp.Kvs {
		var obj cmdb.Object
		err = decode(kvs, &obj)
		if err != nil {
			return err
		}
		if len(opts.LabelSelector) != 0 && !mactchSelector(opts.LabelSelector, obj.GetMeta().Labels) {
			continue
		}
		// TODO: field selector
		*out = append(*out, obj)
	}
	return nil
}

func (s *Store) Create(ctx context.Context, obj cmdb.Object, out *cmdb.Object) error {
	kind := obj.GetKind()
	meta := obj.GetMeta()
	key := s.getStoragePath(obj)

	if err := runtime.ValidateObject(obj); err != nil {
		return NewInvalidObjError(key, err.Error())
	}

	if meta.Version != 0 || meta.Revision != 0 || meta.CreateRevision != 0 {
		return fmt.Errorf("resourceVersion should not be set on objects to be created")
	}

	meta.CreationTimeStamp = time.Now()
	meta.ManagedFields.Time = time.Now()
	defaults.SetDefaults(obj)
	data, err := json.Marshal(obj)
	if err != nil {
		return NewInternalError(err.Error())
	}

	if err = s.handleReferences(ctx, obj, referenceActionCheckExist); err != nil {
		return err
	}

	txnResp, err := s.client.KV.Txn(ctx).If(
		notFound(key),
	).Then(
		clientv3.OpPut(key, string(data)),
	).Commit()
	if err != nil {
		return NewInternalError(err.Error())
	}
	if !txnResp.Succeeded {
		return NewKeyExistsError(key, 0)
	}

	if err = s.handleReferences(ctx, obj, referenceActionCreate); err != nil {
		return err
	}

	if out != nil {
		return s.Get(ctx, kind, meta.Name, meta.Namespace, GetOptions{}, out)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, obj cmdb.Object, out *cmdb.Object) error {
	kind := obj.GetKind()
	meta := obj.GetMeta()
	key := s.getStoragePath(obj)

	if err := runtime.ValidateObject(obj); err != nil {
		return NewInvalidObjError(key, err.Error())
	}

	for {
		var originObj cmdb.Object
		if err := s.Get(ctx, kind, meta.Name, meta.Namespace, GetOptions{}, &originObj); err != nil {
			return err
		}

		originData, err := json.Marshal(originObj)
		if err != nil {
			return err
		}
		originMeta := originObj.GetMeta()

		// 系统管理字段使用旧值
		meta.CreateRevision = originMeta.CreateRevision
		meta.Revision = originMeta.Revision
		meta.Version = originMeta.Version
		meta.CreationTimeStamp = originMeta.CreationTimeStamp
		meta.ManagedFields.Time = originMeta.ManagedFields.Time
		defaults.SetDefaults(obj)
		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		// 无变更，直接返回
		if bytes.Equal(originData, data) {
			if out != nil {
				*out = nil
			}
			return nil
		}

		// 更新此次变更的时间
		meta.ManagedFields.Time = time.Now()

		if err = s.handleReferences(ctx, obj, referenceActionCheckExist); err != nil {
			return err
		}

		txnResp, err := s.client.KV.Txn(ctx).If(
			clientv3.Compare(clientv3.ModRevision(key), "=", originMeta.Revision),
		).Then(
			clientv3.OpPut(key, string(data)),
		).Commit()
		if err != nil {
			return NewInternalError(err.Error())
		}
		if !txnResp.Succeeded {
			// Revision 不一致时应重试
			continue
		}

		// 更新关联关系
		if err = s.updateReferences(ctx, obj, originObj); err != nil {
			return err
		}

		if out != nil {
			return s.Get(ctx, kind, meta.Name, meta.Namespace, GetOptions{}, out)
		}
	}
}

func (s *Store) Delete(ctx context.Context, kind, name, namespace string) error {
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
		return NewInternalError(err.Error())
	}
	if err = s.handleReferences(ctx, obj, referenceActionDelete); err != nil {
		return err
	}
	return nil
}

// 获取资源存储路径
func (s *Store) getStoragePath(obj cmdb.Object) string {
	meta := obj.GetMeta()
	key := strings.ToLower(obj.GetKind()) + "s"
	if meta.HasNamespace() {
		key = path.Join(key, meta.Namespace)
	}
	key = path.Join(s.pathPrefix, key, meta.Name)
	return key
}

func (s *Store) getStoragePathPrefix(kind, namespace string, all bool) string {
	key := strings.ToLower(kind) + "s"
	if namespace != "" && !all {
		key = path.Join(s.pathPrefix, key, namespace)
	} else {
		key = path.Join(s.pathPrefix, key)
	}
	key += "/"
	return key
}

// 创建/删除 引用关系，或检查引用的目标对象是否存在
func (s *Store) handleReferences(ctx context.Context, obj cmdb.Object, action referenceAction) error {
	meta := obj.GetMeta()
	key := s.getStoragePath(obj)
	var refCmps []clientv3.Cmp
	var refOps []clientv3.Op
	refs := runtime.GetFieldValueByTag(reflect.ValueOf(obj), "", "reference")
	for _, ref := range refs {
		if ref.FieldValue != "" {
			refKey := path.Join(s.pathPrefix, "references", ref.TagValue, ref.FieldValue, obj.GetKind(), meta.Name)
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
			switch action {
			case referenceActionCreate:
				refOps = append(refOps, clientv3.OpPut(refKey, ""))
			case referenceActionDelete:
				refOps = append(refOps, clientv3.OpDelete(refKey))
			}
		}
	}

	txnResp, err := s.client.KV.Txn(ctx).If(
		refCmps...,
	).Then(
		refOps...,
	).Commit()
	if err != nil {
		return NewInternalError(err.Error())
	}
	if !txnResp.Succeeded {
		return NewReferencedNotExist(key, fmt.Sprintf("reference object does not exists, %v.", refs))
	}
	return nil
}

// 更新引用关系
func (s *Store) updateReferences(ctx context.Context, obj cmdb.Object, originObj cmdb.Object) error {
	if err := s.handleReferences(ctx, obj, referenceActionDelete); err != nil {
		return err
	}
	if err := s.handleReferences(ctx, obj, referenceActionCreate); err != nil {
		return err
	}
	return nil
}

// 检查当前对象是否被引用
func (s *Store) checkExistReferenced(ctx context.Context, obj cmdb.Object) error {
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

func mactchSelector(selector map[string]string, labels map[string]string) bool {
	for lk, lv := range selector {
		if labels[lk] != lv {
			return false
		}
	}
	return true
}
