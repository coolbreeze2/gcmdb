package client

import (
	"goTool/pkg/cmdb"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func testCreateResource(t *testing.T, o cmdb.Resource, filePath string) {
	file, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var r map[string]any
	err = yaml.Unmarshal(file, &r)
	assert.NoError(t, err)

	metadata := r["metadata"].(map[string]any)
	name := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)
	obj, err := DefaultCMDBClient.CreateResource(o, name, namespace, r)
	if err != nil {
		assert.IsType(t, ResourceAlreadyExistError{}, err)
	} else {
		assert.NoError(t, err)
		assert.IsType(t, map[string]any{}, obj)
	}

	_, err = DefaultCMDBClient.CreateResource(o, name, namespace, r)
	assert.IsType(t, ResourceAlreadyExistError{}, err)
}

func testReadResource(t *testing.T, o cmdb.Resource, name, namespace string) {
	obj, err := DefaultCMDBClient.ReadResource(o, name, namespace, 0)
	assert.IsType(t, map[string]any{}, obj)
	assert.NoError(t, err)
}

func testListResource(t *testing.T, o cmdb.Resource) {
	objs, err := DefaultCMDBClient.ListResource(o, &ListOptions{})
	assert.Less(t, 0, len(objs))
	assert.NoError(t, err)
}

func testCountResource(t *testing.T, o cmdb.Resource, namespace string) {
	count, err := DefaultCMDBClient.CountResource(o, namespace)
	assert.LessOrEqual(t, 1, count)
	assert.NoError(t, err)
}

func testGetResourceNames(t *testing.T, o cmdb.Resource, namespace string) {
	names, err := DefaultCMDBClient.GetResourceNames(o, namespace)
	assert.LessOrEqual(t, 1, len(names))
	assert.NoError(t, err)
}

func testUpdateResource(t *testing.T, o cmdb.Resource, name, namespace, updatePath string, value any) {
	obj, err := DefaultCMDBClient.ReadResource(o, name, namespace, 0)
	assert.NoError(t, err)

	if value == nil {
		value = RandomString(6)
	}
	err = SetMapValueByPath(obj, updatePath, value)
	assert.NoError(t, err)

	obj, err = DefaultCMDBClient.UpdateResource(o, name, namespace, obj)
	assert.NoError(t, err)
	assert.Equal(t, value, GetMapValueByPath(obj, updatePath))

	// 重复执行，无变化
	obj2, err := DefaultCMDBClient.UpdateResource(o, name, namespace, obj)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(obj2))
}

func testDeleteResource(t *testing.T, o cmdb.Resource, name, namespace string) {
	obj, err := DefaultCMDBClient.DeleteResource(o, name, namespace)
	assert.Nil(t, obj)
	assert.NoError(t, err)

	_, err = DefaultCMDBClient.DeleteResource(o, name, namespace)
	assert.IsType(t, ResourceNotFoundError{}, err)
}

func TestCreateResource(t *testing.T) {
	type Case struct {
		o        cmdb.Resource
		filePath string
	}
	cases := []Case{
		{cmdb.NewSecret(), "../example/files/secret.yaml"},
		{cmdb.NewDatacenter(), "../example/files/datacenter.yaml"},
		{cmdb.NewSCM(), "../example/files/scm.yaml"},
		{cmdb.NewProject(), "../example/files/project.yaml"},
		{cmdb.NewApp(), "../example/files/app.yaml"},
	}
	for i := range cases {
		testCreateResource(t, cases[i].o, cases[i].filePath)
	}
}

func TestReadResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o               cmdb.Resource
		name, namespace string
	}
	cases := []Case{
		{cmdb.NewSecret(), "test", ""},
		{cmdb.NewDatacenter(), "test", ""},
		{cmdb.NewSCM(), "gitlab-test", ""},
		{cmdb.NewProject(), "go-devops", ""},
		{cmdb.NewApp(), "go-app", ""},
	}
	for i := range cases {
		testReadResource(t, cases[i].o, cases[i].name, cases[i].namespace)
	}
}

func TestListResource(t *testing.T) {
	TestCreateResource(t)
	cases := []cmdb.Resource{
		cmdb.NewSecret(),
		cmdb.NewDatacenter(),
		cmdb.NewSCM(),
		cmdb.NewApp(),
		cmdb.NewProject(),
	}
	for i := range cases {
		testListResource(t, cases[i])
	}
}

func TestCountResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o         cmdb.Resource
		namespace string
	}
	cases := []Case{
		{cmdb.NewSecret(), ""},
		{cmdb.NewDatacenter(), ""},
		{cmdb.NewSCM(), ""},
		{cmdb.NewProject(), ""},
		{cmdb.NewApp(), ""},
	}
	for i := range cases {
		testCountResource(t, cases[i].o, cases[i].namespace)
	}
}

func TestGetResourceNames(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o         cmdb.Resource
		namespace string
	}
	cases := []Case{
		{cmdb.NewSecret(), ""},
		{cmdb.NewDatacenter(), ""},
		{cmdb.NewSCM(), ""},
		{cmdb.NewProject(), ""},
		{cmdb.NewApp(), ""},
	}
	for i := range cases {
		testGetResourceNames(t, cases[i].o, cases[i].namespace)
	}
}

func TestUpdateResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o                           cmdb.Resource
		name, namespace, updatePath string
		value                       any
	}
	cases := []Case{
		{cmdb.NewSecret(), "test", "", "data.privateKey", "MTIzMg=="},
		{cmdb.NewDatacenter(), "test", "", "spec.provider", "huawei-cloud"},
		{cmdb.NewSCM(), "gitlab-test", "", "spec.url", "https://gitlab-sec.dev.com"},
		{cmdb.NewProject(), "go-devops", "", "spec.nameInChain", nil},
		{cmdb.NewApp(), "go-app", "", "spec.scm.user", nil},
	}
	for i := range cases {
		testUpdateResource(
			t, cases[i].o,
			cases[i].name,
			cases[i].namespace,
			cases[i].updatePath,
			cases[i].value,
		)
	}
}

func TestDeleteResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o               cmdb.Resource
		name, namespace string
	}
	// 优先级倒序
	cases := []Case{
		{cmdb.NewApp(), "go-app", ""},
		{cmdb.NewProject(), "go-devops", ""},
		{cmdb.NewSCM(), "gitlab-test", ""},
		{cmdb.NewDatacenter(), "test", ""},
		{cmdb.NewSecret(), "test", ""},
	}
	for i := range cases {
		testDeleteResource(t, cases[i].o, cases[i].name, cases[i].namespace)
	}
}
