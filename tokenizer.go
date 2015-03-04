package curlyconf

import (
	"regexp"
	"io/ioutil"
	"fmt"
	"unicode/utf8"
)

type Pos struct {
	Line	int
	Column	int
	offset	int
}

type Tokdef struct {
	Match	string
	Token	uint64
	re	*regexp.Regexp
}

type Tokenizer struct {
	file	string
	data	[]byte
	pos	Pos
	tokdef	[]*Tokdef
	space	*regexp.Regexp
	comment	uint64
}

type Tokinfo struct {
	Token	uint64
	Value	[]byte
	Pos	Pos
	tkz	*Tokenizer
}

const (
	TokUnknown = 1 << (63 - iota)
	TokEOF
)

const TokAny = 0x00ffffffffffffff

func newTokenizer(data []byte, td []*Tokdef) (l *Tokenizer)  {
	l = &Tokenizer{ 
		file: `[internal]`,
		data: data,
		tokdef: td,
		space: regexp.MustCompile(`^[\r\t ]+`),
		pos: Pos{Line: 1, Column: 1},
	}
	for i := range td {
		td[i].re = regexp.MustCompile("^" + td[i].Match)
	}
	return
}

func NewTokenizerFromString(data string, td []*Tokdef) (l *Tokenizer, err error)  {
	l = newTokenizer([]byte(data), td)
	return
}

func NewTokenizer(fn string, td []*Tokdef) (l *Tokenizer, err error)  {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return
	}
	l = newTokenizer(data, td)
	l.file = fn
	return
}

func (l *Tokenizer) SetPos(t *Tokinfo) {
	l.pos = t.Pos
}

func (l *Tokenizer) SetSpace(spc string) {
	l.space = regexp.MustCompile(`^[` + spc + `]+`)
}

func (l *Tokenizer) IgnoreComments(c uint64) {
	l.comment = c
}

func (l *Tokenizer) updatePos(s []byte) {
	if (s == nil) {
		return
	}
	for i := l.pos.offset; i < l.pos.offset + len(s); i++ {
		l.pos.Column++
		if (l.data[i] == '\n') {
			l.pos.Line++
			l.pos.Column = 1
		}
	}
	l.pos.offset += len(s)
}

func (l *Tokenizer) skipSpace() {
	s := l.space.Find(l.data[l.pos.offset:])
	if s != nil {
		//fmt.Printf("space found len %d\n", len(s))
		l.updatePos(s)
	}
}

func (l *Tokenizer) peek() (t *Tokinfo) {
	l.skipSpace()
	t = &Tokinfo{}
	t.tkz = l
	t.Pos = l.pos
	if (l.pos.offset == len(l.data)) {
		t.Token = TokEOF;
		return
	}

	//fmt.Printf("start at offset %d\n", l.pos.offset)

	t.Token = TokUnknown
	matchlen := -1

	for i := range l.tokdef {
		s := l.tokdef[i].re.Find(l.data[l.pos.offset:])
		if s != nil && len(s) >= matchlen {
			if len(s) == matchlen {
				t.Token |= l.tokdef[i].Token
			} else {
				matchlen = len(s)
				t.Token = l.tokdef[i].Token
				t.Value = s
			}
			//fmt.Printf("peek: %d match: %s\n", matchlen, l.tokdef[i].re.String())
		}
	}
	return
}

func (l *Tokenizer) Peek() (t *Tokinfo) {
	for {
		t = l.peek()
		if (t.Token & l.comment) == 0 {
			break
		}
		l.updatePos(t.Value)
	}
	return
}

func (l *Tokenizer) Next() (t *Tokinfo) {
	t = l.Peek()
	if (t.Token != TokEOF) {
		l.updatePos(t.Value)
	}
	return
}

func (t *Tokinfo) Error(txt string) (ret []string) {

	ret = append(ret, fmt.Sprintf("%s:%d.%d: ",
			t.tkz.file, t.Pos.Line, t.Pos.Column) + txt)
	if t.Token == TokEOF {
		return
	}

	start := t.Pos.offset
	for start > 0 && t.tkz.data[start - 1] != '\n' {
		start--
	}
	end := t.Pos.offset
	for end < len(t.tkz.data) && t.tkz.data[end] != '\n' {
		end++
	}
	ret = append(ret, string(t.tkz.data[start:end]))

	var u string
	p := start
	for p < t.Pos.offset {
		r, len := utf8.DecodeRune(t.tkz.data[p:end])
		if r == '\t' {
			u = u + "\t"
		} else {
			u = u + " "
		}
		p += len
	}
	u += "^"
	for i := 1; i < len(t.Value); i++ {
		u += "~"
	}
	ret = append(ret, u)

	return
}

