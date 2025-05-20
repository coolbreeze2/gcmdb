package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlJoin(t *testing.T) {
	baseUrl := "http://123.com/api/v1"
	expectedUrl := "http://123.com/api/v1/apps/dev-app/"
	url, err := UrlJoin(baseUrl, "apps", "dev-app/")
	assert.Equal(t, expectedUrl, url)
	assert.NoError(t, err)
}

func TestPath(t *testing.T) {
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

func TestPathRootNil(t *testing.T) {
	m := map[string]any{}
	v := GetMapValueByPath(m, "foo.bar.baz.some.other.thing")
	assert.Equal(t, nil, v)
}

func TestPathNil(t *testing.T) {
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
	assert.Error(t, err, MapKeyPathError{path})
}
