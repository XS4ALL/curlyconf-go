# Curlyconf library

Curlyconf is a configuration file reader for the configuration
file format used by, for example, named.conf and dhcpd.conf.

## Example config

	persons {
		person charlie {
			name "Charlie Brown";
			address 192.168.1.1;
		}
		person snoopy {
			name "Snoopy";
		}
	}

## Example code

	import "curlyconf"

	type Tperson struct {
		Name     string
		Address  net.IP
	}

	type Tpersons struct {
		Persons  []Tpersons
	}

	func main {
		var top Persons
		p, err := curlyconf.ConfParser("file.cfg", curlyconf.ParserSemi)
		if err == nil {
			p.Parse(&top)
		}
		if err != nil {
			Fatal(err)
		}
		for i, p := range top.Persons {
			fmt.Printf("%d name %s\n", i, p.Name)
		}
	}

## This will print:

	1. Charlie Brown
	2. Snoopy

Curlyconf works a lot like json.Unmarshal(), see
http://golang.org/pkg/encoding/json/#Unmarshal . It uses reflection
to match the section or field in the configuration file with
a field in a struct in the code.

## Currently supported types

* integers and floats
* strings
* arrays
* net.IP
* time.Duration
* any type that has a Parse(string) (e error) method

If a field is a slice of one of the above types, the value can be a
comma seperated list. The field can also be a pointer to one of the
above types, a value will be allocated and the pointer set to it.


