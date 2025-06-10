package storage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"goTool/global"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/conversion"
	"math/rand"
	"os"
	"testing"
	"time"

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

func testUpdate(t *testing.T, ctx context.Context, s *store, obj cmdb.Object, name, namespace, updatePath string, value any) {
	var mapObj, updatedMapObj map[string]any
	var updatedObj cmdb.Object
	err := s.Get(ctx, obj.GetKind(), name, namespace, GetOptions{}, &obj)

	assert.NoError(t, err)
	meta := obj.GetMeta()

	if value == nil {
		value = RandomString(6)
	}

	err = conversion.StructToMap(obj, &mapObj)
	oldVal := conversion.GetMapValueByPath(mapObj, updatePath)
	err = conversion.SetMapValueByPath(mapObj, updatePath, value)
	assert.NoError(t, err)

	jsonByte, err := json.Marshal(mapObj)
	assert.NoError(t, err)

	updatedObj, err = conversion.DecodeObject(jsonByte)
	assert.NoError(t, err)

	updatedObj = removeResourceManageFields(updatedObj)
	err = s.Update(ctx, updatedObj, &updatedObj)
	assert.NoError(t, err)

	err = conversion.StructToMap(updatedObj, &updatedMapObj)
	assert.NoError(t, err)

	newValue := conversion.GetMapValueByPath(updatedMapObj, updatePath)
	if oldVal == value {
		assert.Equal(t, oldVal, newValue)
	} else {
		assert.Equal(t, value, newValue, fmt.Sprintf("%s %s %s", obj.GetKind(), meta.Name, meta.Namespace))
	}

	// 重复执行，无变化
	var updatedObj2 cmdb.Object
	updatedObj = removeResourceManageFields(updatedObj)
	err = s.Update(ctx, updatedObj, &updatedObj2)
	assert.NoError(t, err)
	assert.Equal(t, updatedObj, updatedObj2)
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

func TestUpdateResource(t *testing.T) {
	TestCreate(t)
	ctx, s, _ := testSetup()

	type Case struct {
		o                           cmdb.Object
		name, namespace, updatePath string
		value                       any
	}
	cases := []Case{
		{cmdb.NewSecret(), "test", "", "data.privateKey", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewDatacenter(), "test", "", "spec.provider", "huawei-cloud"},
		{cmdb.NewZone(), "test", "", "spec.provider", "huawei-cloud"},
		{cmdb.NewNamespace(), "test", "", "spec.bizEnv", RandomString(6)},
		{cmdb.NewDeployTemplate(), "docker-compose-test", "test", "spec.deployArgs", RandomString(6)},
		{cmdb.NewSCM(), "gitlab-test", "", "spec.url", "https://" + RandomString(6)},
		{cmdb.NewHostNode(), "test", "", "spec.id", RandomString(22)},
		{cmdb.NewHelmRepository(), "test", "", "spec.auth", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewContainerRegistry(), "harbor-test", "", "spec.auth.password", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewConfigCenter(), "apollo-test", "", "spec.apollo.auth", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewDeployPlatform(), "test", "", "spec.kubernetes.cluster.ca", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewProject(), "go-devops", "", "spec.nameInChain", nil},
		{cmdb.NewApp(), "go-app", "", "spec.scm.user", nil},
		{cmdb.NewResourceRange(), "test", "test", "description", RandomString(6)},
		{cmdb.NewOrchestration(), "test", "", "description", RandomString(6)},
		{cmdb.NewAppDeployment(), "go-app", "test", "description", RandomString(6)},
	}
	for i := range cases {
		testUpdate(
			t, ctx, s,
			cases[i].o,
			cases[i].name,
			cases[i].namespace,
			cases[i].updatePath,
			cases[i].value,
		)
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

// 生成随机字符串
func RandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// 移除系统管理字段
func removeResourceManageFields(o cmdb.Object) cmdb.Object {
	meta := o.GetMeta()
	meta.CreateRevision = 0
	meta.Revision = 0
	meta.Version = 0
	return o
}
