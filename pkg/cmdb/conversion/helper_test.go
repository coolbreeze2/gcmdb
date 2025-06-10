package conversion

import (
	"encoding/json"
	"goTool/pkg/cmdb"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func TestInvalidPtrValueKind(t *testing.T) {
	var simple any
	switch obj := simple.(type) {
	default:
		_, err := EnforcePtr(obj)
		if err == nil {
			t.Errorf("Expected error on invalid kind")
		}
	}
}

func TestInvalidMapValueKind(t *testing.T) {
	simple := map[string]string{}
	_, err := EnforcePtr(simple)
	if err == nil {
		t.Errorf("Expected error on invalid kind")
	}
}

func TestEnforceNilPtr(t *testing.T) {
	var nilPtr *struct{}
	_, err := EnforcePtr(nilPtr)
	if err == nil {
		t.Errorf("Expected error on nil pointer")
	}
}

func TestParseResourceFromByteYamlInvalid(t *testing.T) {
	_, err := DecodeObject([]byte("{]"))
	assert.IsType(t, &yaml.SyntaxError{}, err)
}

func TestParseResourceFromByteKindInvalid(t *testing.T) {
	_, err := DecodeObject([]byte("kind: a-not-exist-kind"))
	assert.IsType(t, cmdb.ResourceTypeError{}, err)
}

func TestParseResourceFromByteUnknownField(t *testing.T) {
	_, err := DecodeObject([]byte(`apiVersion: v1alpha
kind: Secret
metadata:
  name: test
  extFiled: v111
data:
  privateKey: 'dGhpcyBpcyBhIHByaXZhdGVLZXkK'`))
	assert.IsType(t, &yaml.UnknownFieldError{}, err)
}

func TestStructToMapUnsupportedType(t *testing.T) {
	var out map[string]any
	ch := make(chan int)
	err := StructToMap(ch, &out)
	assert.IsType(t, &json.UnsupportedTypeError{}, err)
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
