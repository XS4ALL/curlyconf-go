//
//	Tests. We have embedded configs in different formats
//	here, which we read into structs, and then check
//	if the content is as expected.
//
package curlyconf

import (
	"fmt"
	"net"
	"testing"
)

type Attr int

type File struct {
	Name_	string
	Dir	string	`cc:"folder,directory"`
	Attr	[]Attr
	Ptr	*string
}

type Main struct {
	File		[]File
	Net		[]net.IPNet
}

func (a *Attr) UnmarshalText(s []byte) (err error) {
	switch string(s) {
		case "v1":
			*a = 1
		case "v2":
			*a = 2
		default:
			err = fmt.Errorf("unknown attr value")
	}
	return
}

var conf1 string = `
file file1 {
	dir	/var/tmp;
	attr	v1,
		v2;
}

file file1 ptr "Hello World";

file file2 {
	directory /var/tmp;
}

net 2001:888:4::42:7d/120;
net 194.109.6.66/32;
`

var conf2 string = `
file file1
  dir /var/tmp
  attr v1,v2
  ptr "Hello World"
end
file file2
  directory /var/tmp
end
`

func testconf(t *testing.T, data string, how int) {
        var top Main
        p, err := NewParserFromString(data, how)
        if err == nil {
                err = p.Parse(&top)
        }
        if err != nil {
		if e2, ok := err.(*ParseError); ok {
                	for _, m := range e2.Detail {
                       		t.Logf("E: %s\n", m)
                	}
		} else {
                       		t.Logf("E: %s\n", err)
		}
		t.Fail()

        }
	if len(top.File) != 2 {
		t.Error("expected 2 file entries")
	}
	if top.File[0].Name_ != "file1" {
		t.Error("file1.name != file1")
	}
	if top.File[0].Dir != "/var/tmp" {
		t.Error("file1.dir != /var/tmp")
	}
	if top.File[1].Name_ != "file2" {
		t.Error("file2.name != file2")
	}
	if top.File[1].Dir != "/var/tmp" {
		t.Error("file2.dir != /var/tmp")
	}
	if top.File[0].Attr[0] != 1 {
		t.Error("file1.attr != v1")
	}
	if top.File[0].Attr[1] != 2 {
		t.Error("file1.attr != v1")
	}
	if top.File[0].Ptr == nil {
		t.Error("file1.ptr == nil")
	}
	if *top.File[0].Ptr != `Hello World` {
		t.Error("file1.ptr != Hello World")
	}
}

func TestConf1(t *testing.T) {
	testconf(t, conf1, ParserSemi)
}

func TestConf2(t *testing.T) {
	testconf(t, conf2, ParserDiablo)
}

