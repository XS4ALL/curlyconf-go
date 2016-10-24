# Curlyconf library

Curlyconf is a configuration file reader for the configuration
file format used by, for example, named.conf and dhcpd.conf.

## Example config (file.cfg)

	person charlie {
		fullname "Charlie Brown";
		address 192.168.1.1;
	}
	person snoopy {
		fullname "Snoopy";
	}
	person snoopy address 5.6.7.8;

## Example code

	package main

	import (
		"fmt"
		"net"
		"github.com/XS4ALL/curlyconf-go"
	)

	type cfgPerson struct {
		Name_	 string
		Fullname string
		Address	 net.IPAddr
	}

	type cfgMain struct {
		Person	 []cfgPerson
	}

	func main() {
		var top cfgMain
		p, err := curlyconf.NewParser("file.cfg", curlyconf.ParserSemi)
		if err == nil {
			err = p.Parse(&top)
		}
		if err != nil {
			fmt.Println(err.(*curlyconf.ParseError).LongError())
			return
		}
		for i, n := range top.Person {
			fmt.Printf("%d: %s fullname %s addr %v\n",
				i, n.Name_, n.Fullname, n.Address)
		}
	}

## This will print:

	0: charlie fullname "Charlie Brown" addr {192.168.1.1 }
	1: snoopy fullname "Snoopy" addr {5.6.7.8 }

Curlyconf works a lot like json.Unmarshal(), see
http://golang.org/pkg/encoding/json/#Unmarshal . It uses reflection
to match the section or field in the configuration file with
a field in a struct in the code.

## Currently supported types

* integers and floats
* strings
* arrays
* net.IPAddr
* net.TCPAddr
* time.Duration
* any type that complies with the encoding.TextUnmarshaler interface.

If a field is a slice of one of the above types, the value can be a
comma seperated list. The field can also be a pointer to one of the
above types, a value will be allocated and the pointer set to it.

## Sections and structs

Sections in the config file correspond to structs in the code.
The main config is always a struct, see cfgMain in the example above.

A struct can contain values and other structs. It can container
pointers to values/structs - if the pointer is nil, that
configuration option was not set. It can also contain slices
of values/structs or pointers to those.

Sections can be "flattened"- as in the example above,

	person snoopy {
		fullname "Snoopy";
	}

is the same as

	person snoopy fullname "Snoopy";

