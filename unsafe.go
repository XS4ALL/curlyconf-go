package curlyconf

import (
	"reflect"
	"unsafe"
)

//
// Jump through unsafe hoops to get pointer. Yuck.
// Why is there no reflect.ToPtr(Value) ?
//
func toPtr(val reflect.Value) reflect.Value {
        p := unsafe.Pointer(val.Addr().Pointer())
        return reflect.NewAt(val.Type(), p)
}

