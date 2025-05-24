package client

import (
	"encoding/json"
	"goTool/pkg/cmdb"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testCreateResource(t *testing.T, filePath string) {
	r, err := ParseResourceFromFile(filePath)
	assert.NoError(t, err)
	obj, err := DefaultCMDBClient.CreateResource(r)
	if err != nil {
		assert.IsType(t, cmdb.ResourceAlreadyExistError{}, err)
	} else {
		assert.NoError(t, err)
		assert.IsType(t, map[string]any{}, obj)
	}

	_, err = DefaultCMDBClient.CreateResource(r)
	assert.IsType(t, cmdb.ResourceAlreadyExistError{}, err)
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
	oldVal := GetMapValueByPath(obj, updatePath)
	assert.NoError(t, err)

	if value == nil {
		value = RandomString(6)
	}

	err = SetMapValueByPath(obj, updatePath, value)
	assert.NoError(t, err)

	jsonByte, err := json.Marshal(obj)
	assert.NoError(t, err)
	err = json.Unmarshal(jsonByte, &o)
	assert.NoError(t, err)

	obj1, err := DefaultCMDBClient.UpdateResource(o)
	assert.NoError(t, err)
	newValue := GetMapValueByPath(obj1, updatePath)
	assert.NoError(t, err)
	if oldVal == value {
		assert.Equal(t, nil, newValue)
	} else {
		assert.Equal(t, value, newValue)
	}

	// 重复执行，无变化
	obj2, err := DefaultCMDBClient.UpdateResource(o)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(obj2))
}

func testDeleteResource(t *testing.T, o cmdb.Resource, name, namespace string) {
	err := DefaultCMDBClient.DeleteResource(o, name, namespace)
	assert.NoError(t, err)

	err = DefaultCMDBClient.DeleteResource(o, name, namespace)
	assert.IsType(t, cmdb.ResourceNotFoundError{}, err)
}

func TestCreateResource(t *testing.T) {
	cases := []string{
		"../example/files/secret.yaml",
		"../example/files/datacenter.yaml",
		"../example/files/scm.yaml",
		"../example/files/project.yaml",
		"../example/files/app.yaml",
	}
	for i := range cases {
		testCreateResource(t, cases[i])
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
		// {cmdb.NewSecret(), "test", "", "data.privateKey", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewDatacenter(), "test", "", "spec.provider", "huawei-cloud"},
		{cmdb.NewSCM(), "gitlab-test", "", "spec.url", "https://" + RandomString(6)},
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
