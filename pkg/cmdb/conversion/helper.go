package conversion

import (
	"fmt"
	"goTool/pkg/cmdb"
	"reflect"

	"github.com/goccy/go-yaml"
)

// EnforcePtr ensures that obj is a pointer of some sort. Returns a reflect.Value
// of the dereferenced pointer, ensuring that it is settable/addressable.
// Returns an error if this is not possible.
func EnforcePtr(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got nil")
	}
	return v.Elem(), nil
}

func DecodeObject(b []byte) (cmdb.Object, error) {
	var r cmdb.Object
	var err error
	var jsonObj map[string]any
	if err = yaml.Unmarshal(b, &jsonObj); err != nil {
		return nil, err
	}
	kind := jsonObj["kind"].(string)

	if r, err = cmdb.NewResourceWithKind(kind); err != nil {
		return nil, err
	}

	// 不允许设置额外字段
	if err := yaml.UnmarshalWithOptions(b, r, yaml.DisallowUnknownField()); err != nil {
		return nil, err
	}

	return r, nil
}
