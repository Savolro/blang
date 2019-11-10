package lexicon

import (
	"fmt"
)

// State defines current state of lexer
type State int

const (
	Start State = iota
	Ident
	Begin
	End
	Slash
	Comment
	Equal
	Int
	Float
	OpBraceO
	OpBraceC
	AndOr
	Eof
)

// Lexer defines an entity which determines lexicon elements
type Lexer struct {
	keywords         map[string]string
	state            State
	offset           int
	buffer           []byte
	line             int
	column           int
	idents           int
	identLevel       int
	tokens           []Token
	tokenStart       int
	tokenStartLine   int
	tokenStartColumn int
}

// Error defines a lexer error
type Error struct {
	Line    int
	Column  int
	Message string
}

// NewLexer is a default constructor for lexer
func NewLexer() *Lexer {
	return &Lexer{
		line:   1,
		column: 1,
		keywords: map[string]string{
			"if":     "KW_IF",
			"elif":   "KW_ELIF",
			"else":   "KW_ELSE",
			"type":   "KW_TYPE",
			"int":    "KW_INT",
			"string": "KW_STRING",
			"float":  "KW_FLOAT",
			"bool":   "KW_BOOL",
			"fn":     "KW_FN",
			"return": "KW_RETURN",
		},
	}
}

// GetTokens converts input
func (l *Lexer) GetTokens(input []byte) (tokens []Token, err error) {
	for l.offset = 0; l.offset < len(input); l.offset++ {
		err = l.lexChar(input[l.offset])
		if err != nil {
			return nil, err
		}
	}
	_ = l.lexChar(' ')
	l.beginToken(Eof)
	l.completeToken("EOF", true)
	return l.tokens, nil
}

func (l *Lexer) beginToken(state State) {
	l.state = state
	l.tokenStart = l.offset
	l.tokenStartLine = l.line
	l.tokenStartColumn = l.column
}

func (l *Lexer) completeToken(tokenType string, advance bool) {
	if !advance {
		l.offset--
		l.column--
	}
	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  string(l.buffer),
		Line:   l.tokenStartLine,
		Column: l.tokenStartColumn,
	})
	l.buffer = []byte{}
	l.state = Start
}

func (l *Lexer) add(c byte) {
	l.buffer = append(l.buffer, c)
}

func (l *Lexer) lexStart(c byte) error {
	switch {
	case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_':
		l.beginToken(Ident)
		l.add(c)
	case c == ' ':
	case c == '\n':
		l.line++
		l.column = 0
		l.idents = 0
		l.beginToken(End)
	case c == ':':
		l.beginToken(Begin)
		l.completeToken("BEGIN", true)
		l.identLevel++
	case c == '/':
		l.beginToken(Slash)
	case c == '=':
		l.beginToken(Equal)
	case c == '&' || c == '|':
		l.beginToken(AndOr)
		l.add(c)
	case c >= '0' && c <= '9':
		l.beginToken(Int)
		l.add(c)
	case c == '.':
		l.beginToken(Float)
		l.add(c)
	case c == '(':
		l.beginToken(OpBraceO)
		l.completeToken("OP_BRACE_O", true)
	case c == ')':
		l.beginToken(OpBraceC)
		l.completeToken("OP_BRACE_C", true)
	default:
		return l.Error(c)
	}
	return nil
}

func (l *Lexer) Error(c byte) error {
	return fmt.Errorf("%d:%d: unexpected symbol: %c", l.line, l.column, c)

}

func (l *Lexer) lexEnd(c byte) error {
	if c == '\t' {
		l.idents++
	} else if l.idents < l.identLevel {
		l.completeToken("END", false)
		l.identLevel--
	} else {
		l.buffer = []byte{}
		l.offset--
		l.column--
		l.state = Start
	}
	return nil
}

func (l *Lexer) lexIdent(c byte) error {
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || (c >= '0' && c <= '9') {
		l.add(c)
	} else if kwType := l.keywords[string(l.buffer)]; kwType != "" {
		l.buffer = []byte{}
		l.completeToken(kwType, false)

	} else {
		l.completeToken("IDENT", false)
	}
	return nil
}

func (l *Lexer) lexSlash(c byte) (err error) {
	if c == '/' {
		l.state = Comment
	} else {
		l.completeToken("OP_DIV", true)
	}
	return nil
}

func (l *Lexer) lexAndOr(c byte) (err error) {
	if c != l.buffer[0] {
		return l.Error(c)
	}
	l.buffer = []byte{}
	if c == '&' {
		l.completeToken("OP_AND", true)
	} else if c == '|' {
		l.completeToken("OP_OR", true)
	} else {
		return l.Error(c)
	}
	return nil
}

func (l *Lexer) lexEqual(c byte) (err error) {
	if c != '=' {
		l.completeToken("OP_ASSIGN", true)
	} else {
		l.completeToken("OP_EQUAL", true)
	}
	return nil
}

func (l *Lexer) lexComment(c byte) (err error) {
	if c != '\n' {
		l.add(c)
	} else {
		l.completeToken("COMMENT", false)
	}
	return nil
}

func (l *Lexer) lexInt(c byte) (err error) {
	if c >= '0' && c <= '9' {
		l.add(c)
	} else if c == '.' {
		l.add(c)
		l.state = Float
	} else {
		l.completeToken("LEX_INT", false)
	}
	return nil
}

func (l *Lexer) lexFloat(c byte) (err error) {
	if c >= '0' && c <= '9' {
		l.add(c)
	} else {
		l.completeToken("LEX_FLOAT", false)
	}
	return nil
}

func (l *Lexer) lexChar(c byte) (err error) {
	switch l.state {
	case Start:
		err = l.lexStart(c)
	case Ident:
		err = l.lexIdent(c)
	case Slash:
		err = l.lexSlash(c)
	case Equal:
		err = l.lexEqual(c)
	case Comment:
		err = l.lexComment(c)
	case Int:
		err = l.lexInt(c)
	case Float:
		err = l.lexFloat(c)
	case End:
		err = l.lexEnd(c)
	}
	if err != nil {
		return err
	}
	l.column++
	return nil
}
