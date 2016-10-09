package curlyconf

import (
	"fmt"
	"net"
)

type cfgPerson struct {
	Name_	 string
	Fullname string
	Address	 net.IPAddr
}

type cfgMain struct {
	Person	 []cfgPerson
}

func Example() {
	var top cfgMain
	p, err := NewParser("file.cfg", ParserSemi)
	if err == nil {
		err = p.Parse(&top)
	}
	if err != nil {
		fmt.Println(err.(*ParseError).LongError())
		return
	}
	for i, n := range top.Person {
		fmt.Printf("%d: %s fullname %s addr %v\n",
			i, n.Name_, n.Fullname, n.Address)
	}
}
