package client

import (
	"encoding/json"
	"goTool/pkg/cmdb"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructToMapUnsupportedType(t *testing.T) {
	var out map[string]any
	ch := make(chan int)
	err := StructToMap(ch, &out)
	assert.IsType(t, &json.UnsupportedTypeError{}, err)
}

func TestUrlJoin(t *testing.T) {
	baseUrl := "http://123.com/api/v1"
	expectedUrl := "http://123.com/api/v1/apps/dev-app/"
	url := UrlJoin(baseUrl, "apps", "dev-app/")
	assert.Equal(t, expectedUrl, url)
}

func TestGetMapValueByPath(t *testing.T) {
	m := map[string]any{
		"foo": map[string]any{
			"bar": map[string]any{
				"baz": "hello",
			},
		},
	}

	v := GetMapValueByPath(m, "foo.bar.baz")
	assert.Equal(t, "hello", v)
}

func TestGetMapValueByPathRootNil(t *testing.T) {
	m := map[string]any{}
	v := GetMapValueByPath(m, "foo.bar.baz.some.other.thing")
	assert.Equal(t, nil, v)
}

func TestGetMapValueByPathNil(t *testing.T) {
	m := map[string]any{
		"foo": map[string]any{
			"bar": map[string]any{
				"baz": "hello",
			},
		},
	}

	v := GetMapValueByPath(m, "foo.bar.baz.some.other.thing")
	assert.Equal(t, nil, v)
}

func TestSetMapValueByPath(t *testing.T) {
	m := map[string]any{
		"foo": map[string]any{
			"bar": map[string]any{
				"baz": "hello",
			},
		},
	}

	err := SetMapValueByPath(m, "foo.bar.baz", "new-world")
	assert.NoError(t, err)
	assert.Equal(t, "new-world", GetMapValueByPath(m, "foo.bar.baz"))
}

func TestSetMapValueByPathRootNil(t *testing.T) {
	m := map[string]any{}
	path := "foo.bar.baz.some.other.thing"
	err := SetMapValueByPath(m, path, "")
	assert.EqualError(t, err, cmdb.MapKeyPathError{KeyPath: path}.Error())
}
