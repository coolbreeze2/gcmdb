package client

import (
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func testCreateResource(t *testing.T, o Object, filePath string) {
	file, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var r map[string]any
	err = yaml.Unmarshal(file, &r)
	assert.NoError(t, err)

	metadata := r["metadata"].(map[string]any)
	name := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)
	obj, err := o.Create(name, namespace, r)
	if err != nil {
		assert.IsType(t, ResourceAlreadyExistError{}, err)
	} else {
		assert.NoError(t, err)
		assert.IsType(t, map[string]any{}, obj)
	}

	_, err = o.Create(name, namespace, r)
	assert.IsType(t, ResourceAlreadyExistError{}, err)
}

func testReadResource(t *testing.T, o Object, name, namespace string) {
	obj, err := o.Read(name, namespace, 0)
	assert.IsType(t, map[string]any{}, obj)
	assert.NoError(t, err)
}

func testListResource(t *testing.T, o Object) {
	objs, err := o.List(&ListOptions{})
	assert.Less(t, 0, len(objs))
	assert.NoError(t, err)
}

func testCountResource(t *testing.T, o Object, namespace string) {
	count, err := o.Count(namespace)
	assert.LessOrEqual(t, 1, count)
	assert.NoError(t, err)
}

func testGetResourceNames(t *testing.T, o Object, namespace string) {
	names, err := o.GetNames(namespace)
	assert.LessOrEqual(t, 1, len(names))
	assert.NoError(t, err)
}

func testUpdateResource(t *testing.T, o Object, name, namespace, updatePath string, value any) {
	obj, err := o.Read(name, namespace, 0)
	assert.NoError(t, err)

	if value == nil {
		value = RandomString(6)
	}
	err = SetMapValueByPath(obj, updatePath, value)
	assert.NoError(t, err)

	obj, err = o.Update(name, namespace, obj)
	assert.NoError(t, err)
	assert.Equal(t, value, GetMapValueByPath(obj, updatePath))

	// 重复执行，无变化
	obj2, err := o.Update(name, namespace, obj)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(obj2))
}

func testDeleteResource(t *testing.T, o Object, name, namespace string) {
	obj, err := o.Delete(name, namespace)
	assert.Nil(t, obj)
	assert.NoError(t, err)

	_, err = o.Delete(name, namespace)
	assert.IsType(t, ResourceNotFoundError{}, err)
}

func TestCreateResource(t *testing.T) {
	type Case struct {
		o        Object
		filePath string
	}
	cases := []Case{
		{NewProject(), "../example/files/project.yaml"},
		{NewApp(), "../example/files/app.yaml"},
	}
	for i := range cases {
		testCreateResource(t, cases[i].o, cases[i].filePath)
	}
}

func TestReadResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o               Object
		name, namespace string
	}
	cases := []Case{
		{NewApp(), "go-app", ""},
		{NewProject(), "go-devops", ""},
	}
	for i := range cases {
		testReadResource(t, cases[i].o, cases[i].name, cases[i].namespace)
	}
}

func TestListResource(t *testing.T) {
	TestCreateResource(t)
	cases := []Object{
		NewProject(),
		NewApp(),
	}
	for i := range cases {
		testListResource(t, cases[i])
	}
}

func TestCountResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o         Object
		namespace string
	}
	cases := []Case{
		{NewProject(), ""},
		{NewApp(), ""},
	}
	for i := range cases {
		testCountResource(t, cases[i].o, cases[i].namespace)
	}
}

func TestGetResourceNames(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o         Object
		namespace string
	}
	cases := []Case{
		{NewProject(), ""},
		{NewApp(), ""},
	}
	for i := range cases {
		testGetResourceNames(t, cases[i].o, cases[i].namespace)
	}
}

func TestUpdateResource(t *testing.T) {
	TestCreateResource(t)
	type Case struct {
		o                           Object
		name, namespace, updatePath string
		value                       any
	}
	cases := []Case{
		{NewProject(), "go-devops", "", "spec.nameInChain", nil},
		{NewApp(), "go-app", "", "spec.scm.user", nil},
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
		o               Object
		name, namespace string
	}
	cases := []Case{
		{NewApp(), "go-app", ""},
		{NewProject(), "go-devops", ""},
	}
	for i := range cases {
		testDeleteResource(t, cases[i].o, cases[i].name, cases[i].namespace)
	}
}
