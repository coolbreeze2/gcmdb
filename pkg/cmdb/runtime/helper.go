package runtime

import (
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/conversion"
	"reflect"
)

// SetZeroValue would set the object of objPtr to zero value of its type.
func SetZeroValue(objPtr cmdb.Object) error {
	v, err := conversion.EnforcePtr(objPtr)
	if err != nil {
		return err
	}
	v.Set(reflect.Zero(v.Type()))
	return nil
}
