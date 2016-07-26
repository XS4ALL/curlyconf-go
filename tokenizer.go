package curlyconf

import (
	"regexp"
	"io/ioutil"
	"fmt"
	"unicode/utf8"
)

type tokPos struct {
	Line	int
	Column	int
	offset	int
}

type tokDef struct {
	Match	string
	Token	uint64
	re	*regexp.Regexp
}

type tokenizer struct {
	file	string
	data	[]byte
	pos	tokPos
	tokdef	[]*tokDef
	space	*regexp.Regexp
	comment	uint64
}

type tokInfo struct {
	Token	uint64
	Value	[]byte
	Pos	tokPos
	tkz	*tokenizer
}

const (
	tokUnknown = 1 << (63 - iota)
	tokEOF
)

const tokAny = 0x00ffffffffffffff

func newtokenizer(data []byte, td []*tokDef) (l *tokenizer)  {
	l = &tokenizer{ 
		file: `[internal]`,
		data: data,
		tokdef: td,
		space: regexp.MustCompile(`^[\r\t ]+`),
		pos: tokPos{Line: 1, Column: 1},
	}
	for i := range td {
		td[i].re = regexp.MustCompile("^" + td[i].Match)
	}
	return
}

func newTokenizerFromString(data string, td []*tokDef) (l *tokenizer, err error)  {
	l = newtokenizer([]byte(data), td)
	return
}

func newTokenizer(fn string, td []*tokDef) (l *tokenizer, err error)  {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return
	}
	l = newtokenizer(data, td)
	l.file = fn
	return
}

func (l *tokenizer) SetPos(t *tokInfo) {
	l.pos = t.Pos
}

func (l *tokenizer) SetSpace(spc string) {
	l.space = regexp.MustCompile(`^[` + spc + `]+`)
}

func (l *tokenizer) IgnoreComments(c uint64) {
	l.comment = c
}

func (l *tokenizer) updatePos(s []byte) {
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

func (l *tokenizer) skipSpace() {
	s := l.space.Find(l.data[l.pos.offset:])
	if s != nil {
		//fmt.Printf("space found len %d\n", len(s))
		l.updatePos(s)
	}
}

func (l *tokenizer) peek() (t *tokInfo) {
	l.skipSpace()
	t = &tokInfo{}
	t.tkz = l
	t.Pos = l.pos
	if (l.pos.offset == len(l.data)) {
		t.Token = tokEOF;
		return
	}

	//fmt.Printf("start at offset %d\n", l.pos.offset)

	t.Token = tokUnknown
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

func (l *tokenizer) Peek() (t *tokInfo) {
	for {
		t = l.peek()
		if (t.Token & l.comment) == 0 {
			break
		}
		l.updatePos(t.Value)
	}
	return
}

func (l *tokenizer) Next() (t *tokInfo) {
	t = l.Peek()
	if (t.Token != tokEOF) {
		l.updatePos(t.Value)
	}
	return
}

func (t *tokInfo) Error(txt string) (ret []string) {

	ret = append(ret, fmt.Sprintf("%s:%d.%d: ",
			t.tkz.file, t.Pos.Line, t.Pos.Column) + txt)
	if t.Token == tokEOF {
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

