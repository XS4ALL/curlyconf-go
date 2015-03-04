//
//	Parse a string and set reflect.Value.
//	Like encoding.TextUnmarshaler, but generic.
//

package curlyconf

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// mappings from suffix to multipliers
var nsf = map[string]uint64{
	"k": 1000,
	"m": 1000000,
	"g": 1000000000,
	"t": 1000000000000,
}

var colonPortRegexp = regexp.MustCompile(`:([0-9]+|[a-z]+[-0-9a-z]*[0-9a-z])$`)

func suffixMult(s string) (r string, m uint64) {
	m = 1
	r = s
	if l := len(s); l > 0 {
		if n, ok := nsf[strings.ToLower(s[l-1:l])]; ok {
			// found suffix
			r = s[0:l-1]
			m = n
		}
	}
	return
}

func convUint(v string, bits int) (i uint64, e error) {
	v, mult := suffixMult(v)
	if i, e = strconv.ParseUint(v, 0, bits); e == nil {
		i *= mult
	}
	return
}

func convInt(v string, bits int) (i int64, e error) {
	v, mult := suffixMult(v)
	if i, e = strconv.ParseInt(v, 0, bits); e == nil {
		i *= int64(mult)
	}
	return
}

func convFloat(v string) (i float64, e error) {
	v, mult := suffixMult(v)
	if i, e = strconv.ParseFloat(v, 0); e == nil {
		i *= float64(mult)
	}
	return
}

func convDuration(v string) (val reflect.Value, e error) {
	var d time.Duration
	if d, e = time.ParseDuration(v); e == nil {
		val = reflect.ValueOf(d)
	}
	return
}

func convTCPAddr(v string) (val reflect.Value, e error) {
	if !colonPortRegexp.MatchString(v) {
		v = v + `:0`
	}
	t, e := net.ResolveTCPAddr("tcp", v)
	if e == nil {
		var obj interface{}
		obj = *t
		val = reflect.ValueOf(obj)
	}
	return
}

func convIP(v string) (val reflect.Value, e error) {
	ip := net.ParseIP(v)
	if ip != nil {
		var obj interface{}
		obj = ip
		val = reflect.ValueOf(obj)
	} else {
		e = fmt.Errorf("not an IP address")
	}
	return
}

func convIPNet(v string) (val reflect.Value, e error) {
	_, net, e := net.ParseCIDR(v)
	if e == nil {
		var obj interface{}
		obj = *net
		val = reflect.ValueOf(obj)
	}
	return
}

//
//	See if this type has a Parse() method.
//
func hasParseInterface(t reflect.Type) bool {
	switch t.Kind() {
		case reflect.Array, reflect.Chan, reflect.Func,
		     reflect.Interface, reflect.Map, reflect.Ptr,
		     reflect.Slice, reflect.UnsafePointer:
		default:
			// convert to pointer
			t = reflect.PtrTo(t)
	}
	_, hasParse := t.MethodByName("Parse")
	return hasParse
}

//
//	If this value has a Parse() method, return the value.
//
func parseInterface(val reflect.Value) (obj interface{ Parse(string) error }) {
	switch val.Type().Kind() {
		case reflect.Array, reflect.Chan, reflect.Func,
		     reflect.Interface, reflect.Map, reflect.Ptr,
		     reflect.Slice, reflect.UnsafePointer:
		default:
			// convert to pointer
			val = toPtr(val)
	}
	_, hasParse := val.Type().MethodByName("Parse")
	if hasParse == true {
		obj = val.Interface().(interface { Parse(string) error })
	}
	return
}

//
//	Set primitive value - bool, int, float, string
//
func setPrimitive(val reflect.Value, s string) (err error) {

	switch val.Type().Kind() {
		case reflect.Bool:
			switch strings.ToLower(s) {
				case "n", "no", "f", "false", "off":
					val.SetBool(false)
				case "y", "yes", "t", "true", "on", "":
					val.SetBool(true)
				default:
					err = fmt.Errorf("not a boolean value")
			}
		case reflect.Uint64:
			var i uint64
			if i, err = convUint(s, 64); err == nil {
				val.SetUint(i)
			}
		case reflect.Int64:
			switch val.Type().String() {
			default:
				var i int64
				if i, err = convInt(s, 64); err == nil {
					val.SetInt(i)
				}
			}
		case reflect.Uint32:
			var i uint64
			if i, err = convUint(s, 32); err == nil {
				val.SetUint(i)
			}
		case reflect.Uint:
			var i uint64
			if i, err = convUint(s, 0); err == nil {
				val.SetUint(i)
			}
		case reflect.Int:
			var i int64
			if i, err = convInt(s, 0); err == nil {
				val.SetInt(i)
			}
		case reflect.Float64:
			var fl float64
			if fl, err = convFloat(s); err == nil {
				val.SetFloat(fl)
			}
		case reflect.String:
			val.SetString(s)
		default:
			err = fmt.Errorf("unsupported type %s",
						val.Type().String())
	}
	return
}

//
//	Set a field to a value.
//
func SetValue(val reflect.Value, s string) (err error) {

	// If the type has a Parse() method, use it.
	obj := parseInterface(val)
	if obj != nil {
		err = obj.Parse(s)
		return
	}

	// Special support for some types
	var newval reflect.Value
	done := true
	switch val.Type().String() {
		case "net.TCPAddr":
			newval, err = convTCPAddr(s)
		case "net.IP":
			newval, err = convIP(s)
		case "net.IPNet":
			newval, err = convIPNet(s)
		case "time.Duration":
			newval, err = convDuration(s)
		default:
			done = false
	}
	if done {
		if err == nil {
			val.Set(newval)
		}
		return
	}

	// Perhaps a primitive type
	err = setPrimitive(val, s)

	return
}

func CanSetValue(t reflect.Type) (r bool) {
	switch t.Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		     reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		     reflect.Uint16, reflect.Uint32, reflect.Uint64,
		     reflect.Float32, reflect.Float64, reflect.String:
			r = true
		case reflect.Struct:
			switch t.String() {
			case "net.IP", "net.IPNet", "net.TCPAddr":
				r = true
			default:
				r = hasParseInterface(t)
			}
	}
	return
}

