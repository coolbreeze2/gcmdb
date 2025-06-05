package storage

import (
	"context"
	"goTool/global"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/conversion"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var cases = []string{
	"../../example/files/secret.yaml",
	"../../example/files/datacenter.yaml",
	"../../example/files/zone.yaml",
	"../../example/files/namespace.yaml",
	"../../example/files/scm.yaml",
	"../../example/files/hostnode.yaml",
	"../../example/files/helm_repository.yaml",
	"../../example/files/container_registry.yaml",
	"../../example/files/config_center.yaml",
	"../../example/files/deploy_platform.yaml",
	"../../example/files/project.yaml",
	"../../example/files/app.yaml",
	"../../example/files/deploy_template.yaml",
	"../../example/files/resource_range.yaml",
	"../../example/files/orchestration.yaml",
	"../../example/files/appdeployment.yaml",
}

func testSetup() (context.Context, *store, *clientv3.Client) {
	endpoint := global.ServerSetting.ETCD_SERVER_HOST + ":" + global.ServerSetting.ETCD_SERVER_PORT
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{endpoint},
	})
	if err != nil {
		panic(err)
	}
	store := New(client, global.StoragePathPrefix)
	ctx := context.Background()
	return ctx, store, client
}

func parseResourceFromFile(filePath string) (cmdb.Object, error) {
	var file []byte
	var err error
	if file, err = os.ReadFile(filePath); err != nil {
		return nil, err
	}

	return conversion.DecodeObject(file)
}

func testCreate(t *testing.T, ctx context.Context, s *store, filePath string) {
	var out cmdb.Object
	obj, err := parseResourceFromFile(filePath)
	assert.NoError(t, err)
	err = s.Create(ctx, obj, &out)
	if err != nil {
		assert.IsType(t, &StorageError{}, err, err.Error())
		assert.Equal(t, err.(*StorageError).Code, ErrCodeKeyExists)
	} else {
		assert.NoError(t, err)
	}

	err = s.Create(ctx, obj, &out)
	assert.IsType(t, &StorageError{}, err, err.Error())
	assert.Equal(t, err.(*StorageError).Code, ErrCodeKeyExists)
}

func testGet(t *testing.T, ctx context.Context, s *store, filePath string) {
	var out cmdb.Object
	obj, err := parseResourceFromFile(filePath)
	meta := obj.GetMeta()
	assert.NoError(t, err)
	err = s.Get(ctx, obj.GetKind(), meta.Name, meta.Namespace, GetOptions{}, &out)
	assert.NoError(t, err)
}

func testDelete(t *testing.T, ctx context.Context, s *store, filePath string) {
	obj, err := parseResourceFromFile(filePath)
	meta := obj.GetMeta()
	assert.NoError(t, err)
	err = s.Delete(ctx, obj.GetKind(), meta.Name, meta.Namespace)
	assert.NoError(t, err)
}

func TestCreate(t *testing.T) {
	ctx, s, _ := testSetup()
	for i := range cases {
		testCreate(t, ctx, s, cases[i])
	}
}

func TestGet(t *testing.T) {
	TestCreate(t)
	ctx, s, _ := testSetup()
	for i := range cases {
		testGet(t, ctx, s, cases[i])
	}
}

func TestDelete(t *testing.T) {
	TestCreate(t)
	ctx, s, _ := testSetup()
	for i := range cases {
		// 倒序删除
		testDelete(t, ctx, s, cases[len(cases)-i-1])
	}
}
