package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gcmdb/global"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/conversion"
	apiv1 "gcmdb/pkg/cmdb/server/apis/v1"
	"gcmdb/pkg/cmdb/server/storage"
	"io/fs"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func testServer() (*httptest.Server, string) {
	r := apiv1.NewRouter(nil)
	ts := httptest.NewServer(r)
	apiUrl := ts.URL + apiv1.PathPrefix
	return ts, apiUrl
}

func testInvalidStoreServer() *httptest.Server {

	client, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{"invalid-endpoint-url"},
	})
	store := storage.New(client, global.StoragePathPrefix)

	r := apiv1.NewRouter(store)

	return httptest.NewServer(r)
}

func testCreateResource(t *testing.T, apiUrl, filePath string) {
	cli := NewCMDBClient(apiUrl)
	r, err := ParseResourceFromFile(filePath)
	assert.NoError(t, err)
	obj, err := cli.CreateResource(r)
	if err != nil {
		assert.IsType(t, cmdb.ResourceAlreadyExistError{}, err, err.Error())
	} else {
		assert.NoError(t, err)
		assert.IsType(t, map[string]any{}, obj)
	}

	_, err = cli.CreateResource(r)
	assert.IsType(t, cmdb.ResourceAlreadyExistError{}, err, err.Error())
}

func testReadResource(t *testing.T, apiUrl string, o cmdb.Object, name, namespace string) {
	cli := NewCMDBClient(apiUrl)
	obj, err := cli.ReadResource(o, name, namespace, 0)
	assert.IsType(t, map[string]any{}, obj)
	assert.NoError(t, err)
}

func testListResource(t *testing.T, apiUrl string, o cmdb.Object, namespace string) {
	cli := NewCMDBClient(apiUrl)
	objs, err := cli.ListResource(o, &ListOptions{Namespace: namespace})
	assert.Less(t, 0, len(objs))
	assert.NoError(t, err)

	// test selector
	objs, err = cli.ListResource(o, &ListOptions{Namespace: namespace, Selector: map[string]string{"x": "y"}})
	assert.LessOrEqual(t, 0, len(objs))
	assert.NoError(t, err)
}

func testCountResource(t *testing.T, apiUrl string, o cmdb.Object, namespace string) {
	cli := NewCMDBClient(apiUrl)
	count, err := cli.CountResource(o, namespace)
	assert.NoError(t, err, o.GetKind())
	assert.LessOrEqual(t, 1, count, o.GetKind())
}

func testGetResourceNames(t *testing.T, apiUrl string, o cmdb.Object, namespace string) {
	cli := NewCMDBClient(apiUrl)
	names, err := cli.GetResourceNames(o, namespace)
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(names), o.GetKind())
}

func testUpdateResource(t *testing.T, apiUrl string, o cmdb.Object, name, namespace, updatePath string, value any) {
	cli := NewCMDBClient(apiUrl)
	obj, err := cli.ReadResource(o, name, namespace, 0)
	oldVal := conversion.GetMapValueByPath(obj, updatePath)
	assert.NoError(t, err)

	if value == nil {
		value = RandomString(6)
	}

	err = conversion.SetMapValueByPath(obj, updatePath, value)
	assert.NoError(t, err)

	jsonByte, err := json.Marshal(obj)
	assert.NoError(t, err)
	err = json.Unmarshal(jsonByte, &o)
	assert.NoError(t, err, string(jsonByte))

	obj1, err := cli.UpdateResource(o)
	assert.NoError(t, err)
	newValue := conversion.GetMapValueByPath(obj1, updatePath)
	assert.NoError(t, err)
	if oldVal == value {
		assert.Equal(t, nil, newValue)
	} else {
		assert.Equal(t, value, newValue, fmt.Sprintf("%s %s %s", o.GetKind(), name, namespace))
	}

	// 重复执行，无变化
	obj2, err := cli.UpdateResource(o)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(obj2))
}

func testDeleteResource(t *testing.T, apiUrl string, o cmdb.Object, name, namespace string) {
	cli := NewCMDBClient(apiUrl)

	err := cli.DeleteResource(o, name, namespace)
	assert.NoError(t, err)

	err = cli.DeleteResource(o, name, namespace)
	assert.IsType(t, cmdb.ResourceNotFoundError{}, err)
}

func TestDefaultClientApiUrl(t *testing.T) {
	cli := DefaultCMDBClient
	url := cli.getCMDBAPIURL()
	assert.IsType(t, "", url)
}

func TestRemoveResourceManageFieldsNil(t *testing.T) {
	var m map[string]any
	RemoveResourceManageFields(m)
	assert.Nil(t, m)
}

func TestParseResourceFromDirNotExist(t *testing.T) {
	_, _, err := ParseResourceFromDir("a-not-exist-dir")
	assert.IsType(t, &fs.PathError{}, err)
}

func TestParseResourceFromFileNotExist(t *testing.T) {
	_, err := ParseResourceFromFile("a-not-exist-file")
	assert.IsType(t, &fs.PathError{}, err)
}

func TestReadResourceWithBadAPIUrl(t *testing.T) {
	cli := NewCMDBClient(":/bad-url.com")
	_, err := cli.ReadResource(cmdb.NewApp(), "go-app", "", 0)
	assert.IsType(t, &url.Error{}, err)
}

func TestHealth(t *testing.T) {
	ts, apiUrl := testServer()
	defer ts.Close()
	cli := NewCMDBClient(apiUrl)

	health := cli.Health()
	assert.Equal(t, true, health)
}

func TestHealthInvalidStorage(t *testing.T) {
	ts := testInvalidStoreServer()
	defer ts.Close()

	apiUrl := ts.URL + apiv1.PathPrefix
	cli := NewCMDBClient(apiUrl)

	health := cli.Health()
	assert.Equal(t, false, health)
}

func TestCreateValidateError(t *testing.T) {
	ts, apiUrl := testServer()
	defer ts.Close()
	cli := NewCMDBClient(apiUrl)

	_, err := cli.CreateResource(cmdb.NewApp())
	assert.IsType(t, cmdb.ResourceValidateError{}, err)
}

func TestCreateValidateServerError(t *testing.T) {
	ts, apiUrl := testServer()
	defer ts.Close()
	cli := NewCMDBClient(apiUrl)

	s := cmdb.NewSecret()
	s.Metadata.Name = "a-test-secret"
	s.Data = map[string]string{"xyz": "111"}
	_, err := cli.CreateResource(s)
	assert.IsType(t, cmdb.ResourceValidateError{}, err)
}

func TestInvalidStorageServerErr(t *testing.T) {
	ts := testInvalidStoreServer()
	defer ts.Close()

	apiUrl := ts.URL + apiv1.PathPrefix
	cli := NewCMDBClient(apiUrl)

	_, err := cli.ReadResource(cmdb.NewApp(), "test", "", 0)
	assert.IsType(t, cmdb.ServerError{}, err)
}

func TestUpdateValidateError(t *testing.T) {
	ts, apiUrl := testServer()
	defer ts.Close()
	cli := NewCMDBClient(apiUrl)

	_, err := cli.UpdateResource(cmdb.NewApp())
	assert.IsType(t, cmdb.ResourceValidateError{}, err)
}

func TestListInvalidateLabelSelector(t *testing.T) {
	ts, apiUrl := testServer()
	defer ts.Close()
	cli := NewCMDBClient(apiUrl)

	objs, err := cli.ListResource(cmdb.NewApp(), &ListOptions{Selector: map[string]string{"x": "1", "y": "="}})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(objs))
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
		"../example/files/deploy_platform.yaml",
		"../example/files/project.yaml",
		"../example/files/app.yaml",
		"../example/files/deploy_template.yaml",
		"../example/files/resource_range.yaml",
		"../example/files/orchestration.yaml",
		"../example/files/appdeployment.yaml",
		"../example/files/appinstance.yaml",
	}
	ts, apiUrl := testServer()
	defer ts.Close()
	for i := range cases {
		testCreateResource(t, apiUrl, cases[i])
	}
}

func TestReadResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o               cmdb.Object
		name, namespace string
	}
	cases := []Case{
		{cmdb.NewSecret(), "test", ""},
		{cmdb.NewDatacenter(), "test", ""},
		{cmdb.NewZone(), "test", ""},
		{cmdb.NewNamespace(), "test", ""},
		{cmdb.NewDeployTemplate(), "docker-compose-test", "test"},
		{cmdb.NewSCM(), "gitlab-test", ""},
		{cmdb.NewHostNode(), "test", ""},
		{cmdb.NewHelmRepository(), "test", ""},
		{cmdb.NewContainerRegistry(), "harbor-test", ""},
		{cmdb.NewConfigCenter(), "apollo-test", ""},
		{cmdb.NewDeployPlatform(), "test", ""},
		{cmdb.NewProject(), "go-devops", ""},
		{cmdb.NewApp(), "go-app", ""},
		{cmdb.NewResourceRange(), "test", "test"},
		{cmdb.NewOrchestration(), "test", ""},
		{cmdb.NewAppDeployment(), "go-app", "test"},
		{cmdb.NewAppInstance(), "go-app--test--eh6hw", "test"},
	}
	ts, apiUrl := testServer()
	defer ts.Close()
	for i := range cases {
		testReadResource(t, apiUrl, cases[i].o, cases[i].name, cases[i].namespace)
	}
}

func TestReadResourceNoNamespace(t *testing.T) {
	ts, apiUrl := testServer()
	defer ts.Close()

	cli := CMDBClient{ApiUrl: apiUrl}
	app := cmdb.NewApp()
	namspace := "not-exist-namespace"
	app.Metadata.Namespace = namspace
	_, err := cli.ReadResource(app, "go-app", namspace, 0)
	assert.ErrorContains(t, err, fmt.Sprintf("%s/apps/ not found at", namspace))
}

func TestListResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o         cmdb.Object
		namespace string
	}
	cases := []Case{
		{cmdb.NewSecret(), ""},
		{cmdb.NewDatacenter(), ""},
		{cmdb.NewZone(), ""},
		{cmdb.NewNamespace(), ""},
		{cmdb.NewDeployTemplate(), "test"},
		{cmdb.NewSCM(), ""},
		{cmdb.NewHostNode(), ""},
		{cmdb.NewHelmRepository(), ""},
		{cmdb.NewContainerRegistry(), ""},
		{cmdb.NewConfigCenter(), ""},
		{cmdb.NewDeployPlatform(), ""},
		{cmdb.NewApp(), ""},
		{cmdb.NewProject(), ""},
		{cmdb.NewResourceRange(), "test"},
		{cmdb.NewOrchestration(), ""},
		{cmdb.NewAppDeployment(), "test"},
		{cmdb.NewAppInstance(), "test"},
	}
	ts, apiUrl := testServer()
	defer ts.Close()
	for i := range cases {
		testListResource(t, apiUrl, cases[i].o, cases[i].namespace)
	}
}

func TestCountResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o         cmdb.Object
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
		{cmdb.NewDeployPlatform(), ""},
		{cmdb.NewProject(), ""},
		{cmdb.NewApp(), ""},
		{cmdb.NewDeployTemplate(), "test"},
		{cmdb.NewResourceRange(), "test"},
		{cmdb.NewOrchestration(), ""},
		{cmdb.NewAppDeployment(), "test"},
		{cmdb.NewAppInstance(), "test"},
	}
	ts, apiUrl := testServer()
	defer ts.Close()
	for i := range cases {
		testCountResource(t, apiUrl, cases[i].o, cases[i].namespace)
	}
}

func TestGetResourceNames(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o         cmdb.Object
		namespace string
	}
	cases := []Case{
		{cmdb.NewSecret(), ""},
		{cmdb.NewDatacenter(), ""},
		{cmdb.NewZone(), ""},
		{cmdb.NewNamespace(), ""},
		{cmdb.NewDeployTemplate(), "test"},
		{cmdb.NewSCM(), ""},
		{cmdb.NewHostNode(), ""},
		{cmdb.NewHelmRepository(), ""},
		{cmdb.NewContainerRegistry(), ""},
		{cmdb.NewConfigCenter(), ""},
		{cmdb.NewDeployPlatform(), ""},
		{cmdb.NewProject(), ""},
		{cmdb.NewApp(), ""},
		{cmdb.NewResourceRange(), "test"},
		{cmdb.NewOrchestration(), ""},
		{cmdb.NewAppDeployment(), "test"},
		{cmdb.NewAppInstance(), "test"},
	}
	ts, apiUrl := testServer()
	defer ts.Close()
	for i := range cases {
		testGetResourceNames(t, apiUrl, cases[i].o, cases[i].namespace)
	}
}

func TestUpdateResource(t *testing.T) {
	TestCreateResource(t)
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
		{cmdb.NewAppInstance(), "go-app--test--eh6hw", "test", "description", RandomString(6)},
	}
	ts, apiUrl := testServer()
	defer ts.Close()
	for i := range cases {
		testUpdateResource(
			t, apiUrl,
			cases[i].o,
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
		o               cmdb.Object
		name, namespace string
	}
	// 优先级倒序
	cases := []Case{
		{cmdb.NewAppInstance(), "go-app--test--eh6hw", "test"},
		{cmdb.NewAppDeployment(), "go-app", "test"},
		{cmdb.NewOrchestration(), "test", ""},
		{cmdb.NewResourceRange(), "test", "test"},
		{cmdb.NewDeployTemplate(), "docker-compose-test", "test"},
		{cmdb.NewApp(), "go-app", ""},
		{cmdb.NewProject(), "go-devops", ""},
		{cmdb.NewDeployPlatform(), "test", ""},
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
	ts, apiUrl := testServer()
	defer ts.Close()
	for i := range cases {
		testDeleteResource(t, apiUrl, cases[i].o, cases[i].name, cases[i].namespace)
	}
}

func TestRenderAppDeployment(t *testing.T) {
	TestCreateResource(t)
	ts, apiUrl := testServer()
	defer ts.Close()

	namespace := "test"
	name := "go-app"
	params := map[string]any{}
	cli := NewCMDBClient(apiUrl)
	result, err := cli.RenderAppDeployment(name, namespace, params)
	assert.NoError(t, err)
	out, _ := yaml.MarshalWithOptions(result, yaml.AutoInt(), yaml.UseLiteralStyleIfMultiline(true))
	fmt.Println(string(out))
}

func TestRenderDeployTemplate(t *testing.T) {
	TestCreateResource(t)
	ts, apiUrl := testServer()
	defer ts.Close()

	namespace := "test"
	name := "go-app"
	params := map[string]any{}
	cli := NewCMDBClient(apiUrl)
	result, err := cli.RenderDeployTemplate(name, namespace, params)
	assert.NoError(t, err)
	out, _ := yaml.MarshalWithOptions(result, yaml.AutoInt(), yaml.UseLiteralStyleIfMultiline(true))
	fmt.Println(string(out))
}
