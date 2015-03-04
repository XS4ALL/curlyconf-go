//
//	Load struct from a config file, using reflection.
//

package curlyconf

import (
	"fmt"
	"reflect"
	"strings"
)

type StructWriter struct {
	stru		reflect.Value
}

type Field struct {
	ident		string
	val		reflect.Value
	elem		reflect.Value
	fieldType	reflect.Type
	elemType	reflect.Type
}

func upperFirst(s string) (r string) {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[0:1]) + strings.ToLower(s[1:])
}

//
//	Constructor for StructWriter
//
func NewStructWriter(obj interface{}) *StructWriter {
	var s StructWriter
	stru := reflect.ValueOf(obj)
	switch stru.Kind() {
	case reflect.Ptr:
        	s.stru = stru.Elem()
	default:
		fmt.Printf("hmz1, %s\n", stru.Type().String())
		panic("NewStructWriter: object is not a pointer-to-struct")
	}
	if s.stru.Kind() != reflect.Struct {
		fmt.Printf("hmz2, %s\n", stru.Type().String())
		panic("NewStructWriter: object is not a pointer-to-struct")
	}
	return &s
}

//
//	Get a description of the field of a struct.
//
func (s *StructWriter) Field(k string) (f *Field, err error) {

	f = &Field{}

	idx := -1
	tp := s.stru.Type()
	var name string
	for i := 0; i < tp.NumField(); i++ {
		// skip if first letter is not uppercase
		sf := tp.Field(i)
		name = sf.Name
		if name[:1] != strings.ToUpper(name[:1]) {
			continue
		}
		// compare fieldname
		name = strings.ToLower(name)
		if name == k {
			idx = i
			break
		}
		// compare tags
		tag := sf.Tag.Get("cc")
		for _, name = range strings.Split(tag, ",") {
			if name == k {
				idx = i
				break
			}
		}
	}

	// Found?
	if idx == -1 {
		err = fmt.Errorf("unknown field %s", k)
		return
	}

	f.val = s.stru.Field(idx)
	if !f.val.CanSet() {
		msg := fmt.Sprintf("field %s of %s is not assignable",
						k, s.stru.Type().Name())
		panic(msg)
	}
	f.ident = name

	f.fieldType = f.val.Type()
	switch f.fieldType.Kind() {
	case reflect.Slice:
		// it's a slice of values.
		f.elemType = f.val.Type().Elem()
		if f.elemType.Kind() == reflect.Ptr {
			panic("no support for slices of pointers to values")
		}
	case reflect.Ptr:
		// pointer to value (at the moment, nil)
		f.elemType = f.val.Type().Elem()
		if f.elemType.Kind() == reflect.Ptr {
			panic("no support for pointers of pointers to values")
		}
	default:
		f.elemType = f.fieldType
	}

	return
}

func (f *Field) IsBool() bool {
	return f.elemType.Kind() == reflect.Bool
}

func (f *Field) IsSlice() bool {
	return f.fieldType.Kind() == reflect.Slice
}

func (f *Field) IsStruct() bool {
	if CanSetValue(f.elemType) {
		return false
	}
	return f.elemType.Kind() == reflect.Struct
}

func (f *Field) HasName() (r bool) {
	if f.elemType.Kind() == reflect.Struct {
	       _, r = f.elemType.FieldByName("Name_")
	}
	return
}

//
//	Set a field to a value.
//
func (f *Field) Set(s string) (err error) {

	// If this is a pointer or a slice, allocate a new Value
	switch f.fieldType.Kind() {
		case reflect.Ptr:
			elemPtr := reflect.New(f.elemType)
			f.val.Set(elemPtr)
			f.elem = reflect.Indirect(elemPtr)
		case reflect.Slice:
			elem := reflect.Indirect(reflect.New(f.elemType))
			f.val.Set(reflect.Append(f.val, elem))
			f.elem = f.val.Index(f.val.Len() - 1)
		default:
			f.elem = f.val
	}

	if f.elemType.Kind() == reflect.Struct && !CanSetValue(f.elemType) {
		// This is a struct. Set the name if we can.
		v := f.elem.FieldByName("Name_")
		if v.IsValid() {
			v.SetString(s)
		}
	} else {
		err = SetValue(f.elem, s)
	}
	return
}

func (f *Field) PtrToElem() interface{} {
	return toPtr(f.elem).Interface()
}

/*
type MyType struct {
	Value	string
}
func (m *MyType) Parse(s string) (err error) {
	m.Value = s
	return
}

type Bar struct {
	Name_	string
	X1	string
	X2	uint64
	X3	[]string
}

type Foo struct {
	Name_	string
	E1	string
	E2	uint64
	E3	[]string
	E4	Bar
	E5	[]Bar
	E6	time.Duration
	E7	MyType
}

func main() {
	f := Foo{}

	sw := NewStructWriter(&f)
	var field *Field
	var err error

	if field, err = sw.Field("e1"); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("e1: %+v\n", field.val)
	field.Set("Hallo")
	field.Set("Daar")

	if field, err = sw.Field("e3"); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("e3: %+v\n", field.val)
	field.Set("Hallo")
	field.Set("Daar")

	if field, err = sw.Field("e6"); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("e6: %+v\n", field.val)
	field.Set("2h")

	if field, err = sw.Field("e7"); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("e7: %+v\n", field.val)
	field.Set("ladida")

	if field, err = sw.Field("e5"); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("e5: %+v\n", field.val)

	if err := field.Set("het-heeft-een-naam"); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	sub := field.PtrToElem()
	fmt.Printf("== %+v\n", sub)
	subsw := NewStructWriter(sub)

	var sfield *Field
	if sfield, err = subsw.Field("X3"); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("x3: %+v\n", sfield.val)
	sfield.Set("Hello")
	sfield.Set("World")


	fmt.Printf("%+v\n", f)
	fmt.Printf("%+v\n", f.E5[0])
}
*/
