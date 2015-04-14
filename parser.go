package curlyconf

import (
	"errors"
        "fmt"
        "strconv"
)

type ParseError struct {
	Detail	[]string
}

func (e *ParseError) Error() string {
	if e == nil || len(e.Detail) == 0 {
		return "curlyconf: unknown empty error";
	}
	return e.Detail[0]
}

type Parser struct {
	tok		*Tokenizer
	stmtEnd		uint64		// \n or ;
	sectionStart	uint64		// { or '\n'
	sectionEnd	uint64		// } or 'end'
	stmtEndStr	string
	sectionStartStr	string
	sectionEndStr	string
	sectionName	string
	errors		ParseError
	errCount	int
	maxErrors	int
}

const (
	ParserSemi = iota
	ParserNL
	ParserDiablo
)

func esc(b []byte) string {
	return strconv.QuoteToASCII(string(b))
}

func debug(format string, a ...interface{}) {
        msg := fmt.Sprintf(format, a...)
	if len(msg) > 0 && msg[len(msg) - 1] != '\n' {
		msg += "\n"
	}
	//fmt.Print(msg)
}

//
//	Add an error to the list of errors.
//
func (p *Parser) error(t *Tokinfo, s string) {
	if p.sectionName != "" {
		s = "section " + p.sectionName + ": " + s
	}
	if t == nil {
		p.errors.Detail = append(p.errors.Detail, s)
		debug("%s", s)
	} else {
		e := t.Error(s)
		p.errors.Detail = append(p.errors.Detail, e...)
		for _, m := range e {
			debug("%s", m)
		}
	}
	p.errCount++
}

//
//	peek() looks for an optional token
//
func (p *Parser) peek(want uint64) (tok *Tokinfo) {
	next := p.tok.Peek()
	if (next.Token & want) != 0 {
		debug("peek: got %s\n", esc(next.Value))
		tok = next
	}
	return
}

//
//	accept() looks for an optional token
//
func (p *Parser) accept(want uint64) (tok *Tokinfo) {
	next := p.tok.Peek()
	if (next.Token & want) != 0 {
		debug("accept: got %s\n", esc(next.Value))
		tok = p.tok.Next()
	}
	return
}

//
//	expect() demands a certain token, otherwise error
//
func (p *Parser) expect(want uint64, ws string) (tok *Tokinfo, match bool) {
	tok = p.tok.Next()
	if tok.Token == TokEOF {
		p.error(tok, "unexpected end-of-file")
		p.errCount = 1000
		return
	}
	if (tok.Token & want) != 0 {
		match = true
		debug("expect: got %s\n", tok.Value)
	} else {
		p.error(tok, "parse error, expected " + ws)
	}
	return
}

//
//	error seen, try to recover.
//
func (p *Parser) recover(tok *Tokinfo) {
	if tok == nil {
		tok = p.tok.Next()
	}
	for {
		if tok.Token == TokEOF {
			return
		}
		if (tok.Token & p.stmtEnd) != 0 {
			return
		}
		if (tok.Token & p.sectionEnd) != 0 {
			if p.stmtEnd != 0 {
				p.tok.SetPos(tok)
			}
			return
		}
		if (tok.Token & p.sectionStart) != 0 {
			tmp := p.stmtEnd
			p.stmtEnd = 0
			p.recover(nil)
			p.stmtEnd = tmp
		}
		tok = p.tok.Next()
	}
}

//
//	New section
//
func (p *Parser) section(sname string, field *Field) {
	var ok bool
	var tok *Tokinfo
	var name string
	var flatmode bool

	debug("section\n")

	if field.HasName() {
		tok, ok = p.expect(TokValue, "section-name")
		if !ok {
			p.recover(tok)
			return
		}
		name = string(tok.Value)
	}

	//
	// flatmode is section { key val; } --> section key val;
	//
	tok = p.peek(TokIdent)
	if tok != nil {
		flatmode = true
	} else {
		tok, ok = p.expect(p.sectionStart, p.sectionStartStr)
		if !ok {
			p.recover(tok)
			return
		}
	}

	oldname := p.sectionName
	p.sectionName = sname

	// New section starts here
	err := field.Section(name)
	if err != nil {
		p.error(tok, err.Error())
		p.recover(tok)
		return
	}

	sw := NewStructWriter(field.PtrToElem())
	if flatmode {
		p.stmt(sw)
		p.accept(p.stmtEnd)
	} else {
		p.stmts(sw, p.sectionEnd)
		if p.sectionEnd == TokEnd {
			p.expect(p.stmtEnd, p.stmtEndStr)
		} else {
			p.accept(p.stmtEnd)
		}
	}

	p.sectionName = oldname
	return
}

//
//	parse a single statement
//
func (p *Parser) stmt(sw *StructWriter) {

	// Empty statements are allowed
	if p.accept(p.stmtEnd) != nil {
		debug("empty stmt\n")
		return
	}

	// Expect identifier
	tok, ok := p.expect(TokIdent, "identifier")
	debug("stmt %s\n", esc(tok.Value))
	if !ok {
		p.recover(tok)
		return
	}

	// See if we known this identifier
	field, err := sw.Field(string(tok.Value))
	if err != nil {
		p.error(tok, err.Error())
		p.recover(tok)
		return
	}

	// It's a section
	if field.IsStruct() {
		p.section(string(tok.Value), field)
		return
	}

	// boolean variables may omit the "true" part
	if field.IsBool() && p.accept(p.stmtEnd) != nil {
		field.Set("true")
		return
	}

	for {
		tok, ok = p.expect(TokValue, "value")
		if !ok {
			break
		}
		err := field.Set(string(tok.Value))
		if err != nil {
			p.error(tok, err.Error())
		}
		if !field.IsSlice() || p.accept(TokComma) == nil {
			tok, ok = p.expect(p.stmtEnd, p.stmtEndStr)
			break
		}
		p.accept(TokNL)
	}

	if !ok {
		p.recover(tok)
	}

	debug("stmt end\n")
}

//
//	Parse a bunch of statements
//
func (p *Parser) stmts(sw *StructWriter, end uint64) {
	for {
		debug("stmts\n")
		if p.accept(end) != nil {
			return
		}
		p.stmt(sw)
		if p.errCount > p.maxErrors {
			break
		}
	}
}

//
//	Start the actual parsing.
//
func (p *Parser) Parse(obj interface{}) (err error) {
	p.stmts(NewStructWriter(obj), TokEOF)
	if p.errCount > 0 {
		if p.errCount > p.maxErrors && p.errCount != 1000 {
			p.error(nil, "too many errors")
		}
		err = &p.errors
	}
	return
}

//
//	Get the full error string
//
func (p *Parser) LongError() error {
	msg := ""
	for i, e := range p.errors.Detail {
		msg = msg + e
		if i < len(p.errors.Detail) - 1 {
			msg += "\n"
		}
	}
	return errors.New(msg)
}

//
//	Return a new Parser object.
//
func newConfParser(src int, data string, how int) (p *Parser, err error) {
	var t *Tokenizer
	var e error
	if src == 0 {
        	t, e = ConfTokenizer(data)
	} else {
        	t, e = ConfTokenizerFromString(data)
	}
        if e != nil {
		err = &ParseError{ Detail: []string{ e.Error() } }
                return
        }
	p = &Parser{
		tok: t,
		stmtEnd: TokSemi,
		stmtEndStr: "';'",
		sectionStart: TokLCBrace,
		sectionStartStr: "'{'",
		sectionEnd: TokRCBrace,
		sectionEndStr: "'}'",
		maxErrors: 10,
	}
	switch how {
		case ParserNL:
			p.stmtEnd = TokNL
			p.stmtEndStr = "newline"
			t.SetSpace(" \t\r")
		case ParserDiablo:
			p.sectionStart = TokNL
			p.sectionStartStr = "newline"
			p.sectionEnd = TokEnd
			p.sectionEndStr = "\"end\""
			p.stmtEnd = TokNL
			p.stmtEndStr = "newline"
			t.SetSpace(" \t\r")
		case ParserSemi:
		default:
	}
	return
}

func ConfParser(file string, how int) (p *Parser, err error) {
	p, err = newConfParser(0, file, how)
	return
}

func ConfParserFromString(data string, how int) (p *Parser, err error) {
	p, err = newConfParser(1, data, how)
	return
}

