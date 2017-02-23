package curlyconf

import (
        "fmt"
        "strconv"
)

// This is returned on error. Detail contains a few lines that
// print the lines that have errors, and pinpoint the location.
type ParseError struct {
	Detail	[]string		// detailed error (line/position)
}

// Returns a short (one-line) error, useful for logs.
func (pe *ParseError) Error() string {
	if pe == nil || len(pe.Detail) == 0 {
		return "curlyconf: unknown empty error";
	}
	return pe.Detail[0]
}

// Returns a multiline error (for printing on tty)
func (pe *ParseError) LongError() string {
	msg := ""
	for i, e := range pe.Detail {
		msg = msg + e
		if i < len(pe.Detail) - 1 {
			msg += "\n"
		}
	}
	return msg
}


type Parser struct {
	tok		*tokenizer
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

// Types of configuration file syntax.
const (
	ParserSemi = iota	// End 'statement' with semicolon
	ParserNL		// End 'statement' with newline
	ParserDiablo		// diablo config file format (deprecated)
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
func (p *Parser) error(t *tokInfo, s string) {
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
func (p *Parser) peek(want uint64) (tok *tokInfo) {
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
func (p *Parser) accept(want uint64) (tok *tokInfo) {
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
func (p *Parser) expect(want uint64, ws string) (tok *tokInfo, match bool) {
	tok = p.tok.Next()
	if tok.Token == tokEOF {
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
func (p *Parser) recover(tok *tokInfo) {
	if tok == nil {
		tok = p.tok.Next()
	}
	for {
		if tok.Token == tokEOF {
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
func (p *Parser) section(sname string, field *structField) {
	var ok bool
	var tok *tokInfo
	var name string
	var flatmode bool

	debug("section\n")

	if field.HasName() {
		tok, ok = p.expect(tokValue, "section-name")
		if !ok {
			p.recover(tok)
			return
		}
		name = string(tok.Value)
		if len(name) > 0 && name[0] == '"' {
			s, err := strconv.Unquote(name)
			if err == nil {
				name = s
			}
		}
	}

	//
	// flatmode is section { key val; } --> section key val;
	//
	tok = p.peek(tokIdent)
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

	sw := newStructWriter(field.PtrToElem())
	if flatmode {
		p.stmt(sw)
		p.accept(p.stmtEnd)
	} else {
		p.stmts(sw, p.sectionEnd)
		if p.sectionEnd == tokEnd {
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
func (p *Parser) stmt(sw *structWriter) {

	// Empty statements are allowed
	if p.accept(p.stmtEnd) != nil {
		debug("empty stmt\n")
		return
	}

	// Expect identifier
	tok, ok := p.expect(tokIdent, "identifier")
	debug("stmt %s\n", esc(tok.Value))
	if !ok {
		p.recover(tok)
		return
	}

	// See if we known this identifier
	field, err := sw.structField(string(tok.Value))
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
		tok, ok = p.expect(tokValue, "value")
		if !ok {
			break
		}
		err := field.Set(string(tok.Value))
		if err != nil {
			p.error(tok, err.Error())
		}
		if !field.IsSlice() || p.accept(tokComma) == nil {
			tok, ok = p.expect(p.stmtEnd, p.stmtEndStr)
			break
		}
		p.accept(tokNL)
	}

	if !ok {
		p.recover(tok)
	}

	debug("stmt end\n")
}

//
//	Parse a bunch of statements
//
func (p *Parser) stmts(sw *structWriter, end uint64) {
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
	p.stmts(newStructWriter(obj), tokEOF)
	if p.errCount > 0 {
		if p.errCount > p.maxErrors && p.errCount != 1000 {
			p.error(nil, "too many errors")
		}
		err = &p.errors
	}
	return
}

//
//	Return a new Parser object.
//
func newConfParser(src int, data string, how int) (p *Parser, err error) {
	var t *tokenizer
	var e error
	if src == 0 {
        	t, e = confTokenizer(data)
	} else {
        	t, e = confTokenizerFromString(data)
	}
        if e != nil {
		err = &ParseError{ Detail: []string{ e.Error() } }
                return
        }
	p = &Parser{
		tok: t,
		stmtEnd: tokSemi,
		stmtEndStr: "';'",
		sectionStart: tokLCBrace,
		sectionStartStr: "'{'",
		sectionEnd: tokRCBrace,
		sectionEndStr: "'}'",
		maxErrors: 10,
	}
	switch how {
		case ParserNL:
			p.stmtEnd = tokNL
			p.stmtEndStr = "newline"
			t.SetSpace(" \t\r")
		case ParserDiablo:
			p.sectionStart = tokNL
			p.sectionStartStr = "newline"
			p.sectionEnd = tokEnd
			p.sectionEndStr = "\"end\""
			p.stmtEnd = tokNL
			p.stmtEndStr = "newline"
			t.SetSpace(" \t\r")
		case ParserSemi:
		default:
	}
	return
}

// Parse a configuration file into Go structures.
// parserType is ParserSemi or ParserNL
//
// The error returned is actually of type ParseError. To get
// at that, use err.(curlyconf.ParseError)
func NewParser(file string, parserType int) (p *Parser, err error) {
	p, err = newConfParser(0, file, parserType)
	return
}

// Like NewParser, but parses a string instead of a file.
func NewParserFromString(data string, parserType int) (p *Parser, err error) {
	p, err = newConfParser(1, data, parserType)
	return
}

