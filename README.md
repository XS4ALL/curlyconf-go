# Curlyconf library

Curlyconf is a configuration file reader for the configuration
file format used by, for example, named.conf and dhcpd.conf.

## License

Copyight (C) XS4ALL Internet bv 2016

Mozilla Public License 2.0

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file,
you can obtain one at http://mozilla.org/MPL/2.0/.

## Example config

	person charlie {
		fullname "Charlie Brown";
		address 192.168.1.1;
	}
	person snoopy {
		fullname "Snoopy";
	}
	person snoopy address 5.6.7.8;

## Example code

	import (
		"fmt"
		"net"
		"github.xs4all.net/XS4ALL/curlyconf-go"
	)

	type cfgPerson struct {
		Name_	 string
		Fullname string
		Address	 net.IPAddr
	}

	type cfgMain struct {
		Person	 []cfgPerson
	}

	func main {
		var top cfgMain
		p, err := curlyconf.ConfParser("file.cfg", curlyconf.ParserSemi)
		if err == nil {
			err = p.Parse(&top)
		}
		if err != nil {
			fmt.Println(p.LongError())
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

T.B.D.

