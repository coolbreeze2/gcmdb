package runtime

import (
	"goTool/pkg/cmdb"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFieldValueByTagEmpty(t *testing.T) {
	app := *cmdb.NewApp()
	v := reflect.ValueOf(app)
	result := GetFieldValueByTag(v, "", "reference")
	assert.Equal(t, []TagValuePair{}, result)
}

func TestGetFieldValueByTag(t *testing.T) {
	app := &cmdb.Datacenter{Spec: cmdb.DatacenterSpec{PrivateKey: "test"}}
	v := reflect.ValueOf(app)
	result := GetFieldValueByTag(v, "", "reference")
	assert.Equal(t, []TagValuePair{{"Secret", "test"}}, result)
}
