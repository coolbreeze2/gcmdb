package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb"
	"io/fs"
	"net/url"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func testCreateResource(t *testing.T, filePath string) {
	r, err := ParseResourceFromFile(filePath)
	assert.NoError(t, err)
	obj, err := DefaultCMDBClient.CreateResource(r)
	if err != nil {
		assert.IsType(t, cmdb.ResourceAlreadyExistError{}, err, err.Error())
	} else {
		assert.NoError(t, err)
		assert.IsType(t, map[string]any{}, obj)
	}

	_, err = DefaultCMDBClient.CreateResource(r)
	assert.IsType(t, cmdb.ResourceAlreadyExistError{}, err, err.Error())
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

	// test selector
	objs, err = DefaultCMDBClient.ListResource(o, &ListOptions{Selector: map[string]string{"x": "y"}})
	assert.LessOrEqual(t, 0, len(objs))
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
		assert.Equal(t, value, newValue, fmt.Sprintf("%s %s %s", o.GetKind(), name, namespace))
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

func TestParseResourceFromDirNotExist(t *testing.T) {
	_, err := ParseResourceFromDir("a-not-exist-dir")
	assert.IsType(t, &fs.PathError{}, err)
}

func TestParseResourceFromFileNotExist(t *testing.T) {
	_, err := ParseResourceFromFile("a-not-exist-file")
	assert.IsType(t, &fs.PathError{}, err)
}

func TestParseResourceFromByteYamlInvalid(t *testing.T) {
	_, err := ParseResourceFromByte([]byte("{]"))
	assert.IsType(t, &yaml.SyntaxError{}, err)
}

func TestParseResourceFromByteKindInvalid(t *testing.T) {
	_, err := ParseResourceFromByte([]byte("kind: a-not-exist-kind"))
	assert.IsType(t, cmdb.ResourceTypeError{}, err)
}

func TestParseResourceFromByteUnknownField(t *testing.T) {
	_, err := ParseResourceFromByte([]byte(`apiVersion: v1alpha
kind: Secret
metadata:
  name: test
  extFiled: v111
data:
  privateKey: 'dGhpcyBpcyBhIHByaXZhdGVLZXkK'`))
	assert.IsType(t, &yaml.UnknownFieldError{}, err)
}

func TestReadResourceWithBadAPIUrl(t *testing.T) {
	cli := NewCMDBClient(":/bad-url.com")
	_, err := cli.ReadResource(cmdb.NewApp(), "go-app", "", 0)
	assert.IsType(t, &url.Error{}, err)
}

func TestCreateValidateError(t *testing.T) {
	_, err := DefaultCMDBClient.CreateResource(cmdb.NewApp())
	assert.IsType(t, validator.ValidationErrors{}, err)
}

func TestCreateValidateServerError(t *testing.T) {
	s := cmdb.NewSecret()
	s.Metadata.Name = "a-test-secret"
	s.Data = map[string]string{"xyz": "111"}
	_, err := DefaultCMDBClient.CreateResource(s)
	assert.IsType(t, cmdb.ResourceValidateError{}, err)
}

func TestUpdateValidateError(t *testing.T) {
	_, err := DefaultCMDBClient.UpdateResource(cmdb.NewApp())
	assert.IsType(t, validator.ValidationErrors{}, err)
}

func TestListInvalidateLabelSelector(t *testing.T) {
	_, err := DefaultCMDBClient.ListResource(cmdb.NewApp(), &ListOptions{Selector: map[string]string{"x": "1", "y": "="}})
	assert.IsType(t, cmdb.ServerError{}, err)
}

func TestCreateResource(t *testing.T) {
	cases := []string{
		"../example/files/secret.yaml",
		"../example/files/datacenter.yaml",
		"../example/files/zone.yaml",
		"../example/files/namespace.yaml",
		"../example/files/scm.yaml",
		"../example/files/hostnode.yaml",
		"../example/files/helm_repository.yaml",
		"../example/files/container_registry.yaml",
		"../example/files/config_center.yaml",
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
		{cmdb.NewZone(), "test", ""},
		{cmdb.NewNamespace(), "test", ""},
		{cmdb.NewSCM(), "gitlab-test", ""},
		{cmdb.NewHostNode(), "test", ""},
		{cmdb.NewHelmRepository(), "test", ""},
		{cmdb.NewContainerRegistry(), "harbor-test", ""},
		{cmdb.NewConfigCenter(), "apollo-test", ""},
		{cmdb.NewProject(), "go-devops", ""},
		{cmdb.NewApp(), "go-app", ""},
		{cmdb.NewZone(), "test", ""},
	}
	for i := range cases {
		testReadResource(t, cases[i].o, cases[i].name, cases[i].namespace)
	}
}

func TestReadResourceNoNamespace(t *testing.T) {
	cli := CMDBClient{}
	app := cmdb.NewApp()
	namspace := "not-exist-namespace"
	app.Metadata.Namespace = namspace
	_, err := cli.ReadResource(app, "go-app", namspace, 0)
	assert.EqualError(t, cmdb.ResourceNotFoundError{Path: cli.getCMDBAPIURL(), Kind: "apps", Namespace: namspace}, err.Error())
}

func TestListResourceNoNamespace(t *testing.T) {
	cli := CMDBClient{}
	app := cmdb.NewApp()
	namspace := "not-exist-namespace"
	app.Metadata.Namespace = namspace
	_, err := cli.ListResource(app, &ListOptions{Namespace: namspace})
	assert.EqualError(t, cmdb.ResourceNotFoundError{Path: cli.getCMDBAPIURL(), Kind: "apps", Namespace: namspace}, err.Error())
}

func TestListResource(t *testing.T) {
	TestCreateResource(t)
	cases := []cmdb.Resource{
		cmdb.NewSecret(),
		cmdb.NewDatacenter(),
		cmdb.NewZone(),
		cmdb.NewNamespace(),
		cmdb.NewSCM(),
		cmdb.NewHostNode(),
		cmdb.NewHelmRepository(),
		cmdb.NewContainerRegistry(),
		cmdb.NewConfigCenter(),
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
		{cmdb.NewZone(), ""},
		{cmdb.NewNamespace(), ""},
		{cmdb.NewSCM(), ""},
		{cmdb.NewHostNode(), ""},
		{cmdb.NewHelmRepository(), ""},
		{cmdb.NewContainerRegistry(), ""},
		{cmdb.NewConfigCenter(), ""},
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
		{cmdb.NewZone(), ""},
		{cmdb.NewNamespace(), ""},
		{cmdb.NewSCM(), ""},
		{cmdb.NewHostNode(), ""},
		{cmdb.NewHelmRepository(), ""},
		{cmdb.NewContainerRegistry(), ""},
		{cmdb.NewConfigCenter(), ""},
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
		{cmdb.NewSecret(), "test", "", "data.privateKey", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewDatacenter(), "test", "", "spec.provider", "huawei-cloud"},
		{cmdb.NewZone(), "test", "", "spec.provider", "huawei-cloud"},
		{cmdb.NewNamespace(), "test", "", "spec.bizEnv", RandomString(6)},
		{cmdb.NewSCM(), "gitlab-test", "", "spec.url", "https://" + RandomString(6)},
		{cmdb.NewHostNode(), "test", "", "spec.id", RandomString(22)},
		{cmdb.NewHelmRepository(), "test", "", "spec.auth", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewContainerRegistry(), "harbor-test", "", "spec.auth.password", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
		{cmdb.NewConfigCenter(), "apollo-test", "", "spec.apollo.auth", base64.StdEncoding.EncodeToString([]byte(RandomString(6)))},
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
		{cmdb.NewConfigCenter(), "apollo-test", ""},
		{cmdb.NewContainerRegistry(), "harbor-test", ""},
		{cmdb.NewHelmRepository(), "test", ""},
		{cmdb.NewHostNode(), "test", ""},
		{cmdb.NewSCM(), "gitlab-test", ""},
		{cmdb.NewZone(), "test", ""},
		{cmdb.NewNamespace(), "test", ""},
		{cmdb.NewDatacenter(), "test", ""},
		{cmdb.NewSecret(), "test", ""},
	}
	for i := range cases {
		testDeleteResource(t, cases[i].o, cases[i].name, cases[i].namespace)
	}
}
