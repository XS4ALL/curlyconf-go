package curlyconf

const Re_ident string = `[a-zA-Z][a-zA-Z0-9-]*[a-zA-Z0-9]*`

const Re_filename string = `\.{0,2}/[0-9a-zA-Z./_-]+`

const Re_dqstring string = `"(?:\\"|[^"])+(:?"|$)`
const Re_bqstring string = "`[^`]*(:?`|$)"

const Re_hostname string = `(?i:([0-9a-z][0-9a-z-]*[0-9a-z]|[0-9a-z]+)` +
			    `(\.([0-9a-z][0-9a-z-]*[0-9a-z]|[0-9a-z]+)+))`;
const Re_hostport string = `(` + Re_hostname + `:\d+)`

const Re_ipv4 string = `(([0-9]{1,3}\.){3}[0-9]{1,3})`
const Re_ipv4port string = `((\*|` + Re_ipv4 + `):\d+)`

const Re_ipv6 string =
	`(?i:(((([0-9a-f]{1,4}:){1,7}|:)(:|(:[0-9a-f]{1,4}){1,7}))|` +
	`([0-9a-f]{1,4}:){7}[0-9a-f]{1,4}))`
const Re_ipv6port string = `(\[` + Re_ipv6 + `\]:\d+)`

const Re_ngmatch string =
	`[@!]?[0-9a-z+_*]+(\.[0-9a-z+_*]+)*`

const Re_comment string = `(//|#)[^\n]*(?:\n|$)`

const (
	TokNL = 1 << iota
	TokLCBrace
	TokRCBrace
	TokLBrace
	TokRBrace
	TokComma
	TokSemi
	TokEqual
	TokInt
	TokFloat
	TokString
	TokIdent
	TokFilename
	TokHostname
	TokHostPort
	TokIP
	TokIPv4
	TokIPv6
	TokIpPort
	TokIPv4Port
	TokIPv6Port
	TokNgMatch
	TokEnd
	TokComment
	TokValue
)

var tokdef = []*Tokdef{
	&Tokdef{ Match: "\n", Token: TokNL },
	&Tokdef{ Match: `{`, Token: TokLCBrace },
	&Tokdef{ Match: `}`, Token: TokRCBrace },
	&Tokdef{ Match: `\(`, Token: TokLBrace },
	&Tokdef{ Match: `\)`, Token: TokRBrace },
	&Tokdef{ Match: `,`, Token: TokComma },
	&Tokdef{ Match: `;`, Token: TokSemi },
	&Tokdef{ Match: `=`, Token: TokEqual },
	&Tokdef{ Match: `\*`, Token: TokIP|TokIPv4|TokValue },
	&Tokdef{ Match: `\d+[kKmMgGtT]`, Token: TokInt|TokValue },
	&Tokdef{ Match: `\d+`, Token: TokInt|TokFloat|TokValue },
	&Tokdef{ Match: `\d+\.\d+`, Token: TokFloat|TokValue },
	&Tokdef{ Match: Re_dqstring, Token: TokString|TokValue },
	&Tokdef{ Match: Re_ident, Token: TokIdent|TokValue },
	&Tokdef{ Match: Re_filename, Token: TokFilename|TokValue },
	&Tokdef{ Match: Re_hostname, Token: TokHostname|TokValue },
	&Tokdef{ Match: Re_hostport, Token: TokHostPort|TokValue },
	&Tokdef{ Match: Re_ipv4,  Token: TokIP|TokIPv4|TokValue },
	&Tokdef{ Match: Re_ipv4port,  Token: TokIpPort|TokIPv4Port|TokValue },
	&Tokdef{ Match: Re_ipv6, Token: TokIP|TokIPv6|TokValue },
	&Tokdef{ Match: Re_ipv6port,  Token: TokIpPort|TokIPv6Port|TokValue },
	&Tokdef{ Match: Re_ngmatch,  Token: TokNgMatch|TokValue },
	&Tokdef{ Match: `end`, Token: TokEnd|TokValue },
	&Tokdef{ Match: `(//|#)[^\n]*(?:\n|$)`, Token: TokComment },
}

func ConfTokenizer(file string) (t *Tokenizer, err error) {
	t, err = NewTokenizer(file, tokdef)
	t.IgnoreComments(TokComment)
	t.SetSpace(" \t\r\n")
	return
}

func ConfTokenizerFromString(data string) (t *Tokenizer, err error) {
	t, err = NewTokenizerFromString(data, tokdef)
	t.IgnoreComments(TokComment)
	t.SetSpace(" \t\r\n")
	return
}

