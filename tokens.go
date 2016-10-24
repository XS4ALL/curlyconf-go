package curlyconf

const re_ident string = `[a-zA-Z][a-zA-Z0-9-]*[a-zA-Z0-9]*`

const re_filename string = `\.{0,2}/[0-9a-zA-Z./_-]+`

const re_dqstring string = `"(?:\\"|[^"])+(:?"|$)`
const re_bqstring string = "`[^`]*(:?`|$)"

const re_hostname string = `(?i:([0-9a-z][0-9a-z-]*[0-9a-z]|[0-9a-z]+)` +
			    `(\.([0-9a-z][0-9a-z-]*[0-9a-z]|[0-9a-z]+)+))`;
const re_hostport string = `(` + re_hostname + `:\d+)`

const re_ipv4 string = `(([0-9]{1,3}\.){3}[0-9]{1,3})(/[0-9]+)?`
const re_ipv4port string = `((\*|` + re_ipv4 + `):\d+)`

const re_ipv6 string =
	`(?i:(((([0-9a-f]{1,4}:){1,7}|:)(:|(:[0-9a-f]{1,4}){1,7}))|` +
	`([0-9a-f]{1,4}:){7}[0-9a-f]{1,4}))(/[0-9]+)?`
const re_ipv6port string = `(\[` + re_ipv6 + `\]:\d+)`

const re_ngmatch string =
	`[@!]?[0-9a-z+_*]+(\.[0-9a-z+_*]+)*`

const re_comment string = `(//|#)[^\n]*(?:\n|$)`

const (
	tokNL = 1 << iota
	tokLCBrace
	tokRCBrace
	tokLBrace
	tokRBrace
	tokComma
	tokSemi
	tokEqual
	tokInt
	tokFloat
	tokString
	tokIdent
	tokFilename
	tokHostname
	tokHostPort
	tokIP
	tokIPv4
	tokIPv6
	tokIpPort
	tokIPv4Port
	tokIPv6Port
	tokNgMatch
	tokEnd
	tokComment
	tokValue
)

var tokdef = []*tokDef{
	&tokDef{ Match: "\n", Token: tokNL },
	&tokDef{ Match: `{`, Token: tokLCBrace },
	&tokDef{ Match: `}`, Token: tokRCBrace },
	&tokDef{ Match: `\(`, Token: tokLBrace },
	&tokDef{ Match: `\)`, Token: tokRBrace },
	&tokDef{ Match: `,`, Token: tokComma },
	&tokDef{ Match: `;`, Token: tokSemi },
	&tokDef{ Match: `=`, Token: tokEqual },
	&tokDef{ Match: `\*`, Token: tokIP|tokIPv4|tokValue },
	&tokDef{ Match: `\d+[kKmMgGtT]`, Token: tokInt|tokValue },
	&tokDef{ Match: `\d+`, Token: tokInt|tokFloat|tokValue },
	&tokDef{ Match: `\d+\.\d+`, Token: tokFloat|tokValue },
	&tokDef{ Match: re_dqstring, Token: tokString|tokValue },
	&tokDef{ Match: re_ident, Token: tokIdent|tokValue },
	&tokDef{ Match: re_filename, Token: tokFilename|tokValue },
	&tokDef{ Match: re_hostname, Token: tokHostname|tokValue },
	&tokDef{ Match: re_hostport, Token: tokHostPort|tokValue },
	&tokDef{ Match: re_ipv4,  Token: tokIP|tokIPv4|tokValue },
	&tokDef{ Match: re_ipv4port,  Token: tokIpPort|tokIPv4Port|tokValue },
	&tokDef{ Match: re_ipv6, Token: tokIP|tokIPv6|tokValue },
	&tokDef{ Match: re_ipv6port,  Token: tokIpPort|tokIPv6Port|tokValue },
	&tokDef{ Match: re_ngmatch,  Token: tokNgMatch|tokValue },
	&tokDef{ Match: `end`, Token: tokEnd|tokValue },
	&tokDef{ Match: `(//|#)[^\n]*(?:\n|$)`, Token: tokComment },
}

func confTokenizer(file string) (t *tokenizer, err error) {
	t, err = newTokenizer(file, tokdef)
	if err == nil {
		t.IgnoreComments(tokComment)
		t.SetSpace(" \t\r\n")
	}
	return
}

func confTokenizerFromString(data string) (t *tokenizer, err error) {
	t, err = newTokenizerFromString(data, tokdef)
	if err == nil {
		t.IgnoreComments(tokComment)
		t.SetSpace(" \t\r\n")
	}
	return
}

