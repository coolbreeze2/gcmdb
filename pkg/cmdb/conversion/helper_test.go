package conversion

import (
	"goTool/pkg/cmdb"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

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
