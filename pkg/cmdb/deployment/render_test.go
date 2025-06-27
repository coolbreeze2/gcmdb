package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToObject_Success(t *testing.T) {
	yamlStr := `apiVersion: v1
kind: AppDeployment
metadata:
  name: test-app
  namespace: test
spec:
  resourceRange: test-range
  template:
    deployTemplate:
      name: test-template
    spec:
      app: test
`
	obj, err := stringToObject(yamlStr)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	meta := obj.GetMeta()
	assert.Equal(t, "test-app", meta.Name)
	assert.Equal(t, "test", meta.Namespace)
}

func TestStringToObject_InvalidYAML(t *testing.T) {
	invalidYAML := `apiVersion: v1
kind: AppDeployment
metadata:
  name: test-app
  namespace: test
spec:
  resourceRange: [unclosed
`
	obj, err := stringToObject(invalidYAML)
	assert.Error(t, err)
	assert.Nil(t, obj)
}

func TestStringToObject_InvalidKind(t *testing.T) {
	// Kind that does not exist in cmdb.Object registry
	yamlStr := `apiVersion: v1
kind: NotExistKind
metadata:
  name: test
`
	obj, err := stringToObject(yamlStr)
	assert.Error(t, err)
	assert.Nil(t, obj)
}
