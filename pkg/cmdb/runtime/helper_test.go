package runtime

import (
	"gcmdb/pkg/cmdb"
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

func TestGetFieldValueByTagMap(t *testing.T) {
	m := map[string]string{}
	v := reflect.ValueOf(m)
	result := GetFieldValueByTag(v, "", "reference")
	assert.Equal(t, []TagValuePair{}, result)
}

func TestGetFieldValueByTagList(t *testing.T) {
	type Case struct {
		Field1 []string `reference:"Secret"`
		Field2 []string `reference:"Secret"`
	}
	c := Case{Field1: []string{"v1", "v1", "v2"}, Field2: []string{"v1", "v1", "v2"}}
	v := reflect.ValueOf(c)
	result := GetFieldValueByTag(v, "", "reference")
	assert.Equal(t, []TagValuePair{{"Secret", "v1"}, {"Secret", "v2"}}, result)
}
