//
//	Parse a string and set reflect.Value.
//	Like encoding.TextUnmarshaler, but generic.
//

package curlyconf

import (
	"encoding"
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
var ipv4Regexp = regexp.MustCompile(`^([0-9]+\.)[0-9]+$`)
var ipv6Regexp = regexp.MustCompile(`:.*:`)

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

func convIPAddr(v string) (val reflect.Value, e error) {
	n := "ip"
	if ipv4Regexp.MatchString(v) {
		n = "ip4"
	}
	if ipv6Regexp.MatchString(v) {
		n = "ip6"
	}
	ip, e := net.ResolveIPAddr(n, v)
	if e == nil {
		var obj interface{}
		obj = *ip
		val = reflect.ValueOf(obj)
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
			if len(s) > 0 && s[0] == '"' {
				s, err = strconv.Unquote(s)
				if err == nil {
					val.SetString(s)
				}
				break
			}
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

	// If the type complies with the TextUnmarshaler interface, use it.
	if val.CanInterface() {
		intf := toPtr(val).Interface()
		if obj, ok := intf.(encoding.TextUnmarshaler); ok {
			err = obj.UnmarshalText([]byte(s))
			return
		}
	}

	// Special support for some types
	var newval reflect.Value
	done := true
	switch val.Type().String() {
		case "net.TCPAddr":
			newval, err = convTCPAddr(s)
		case "net.IPAddr":
			newval, err = convIPAddr(s)
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
			case "net.IPAddr", "net.IPNet", "net.TCPAddr":
				r = true
			default:
				tumtype := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
				r = t.Implements(tumtype)
			}
	}
	return
}

