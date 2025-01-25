package sql_parser

import (
	"fmt"
	"strings"
)

type tokenType string

const (
	tokenType_RESERVED   tokenType = "reserved"
	tokenType_IDENTIFIER tokenType = "identifier"
	tokenType_SYMBOL     tokenType = "symbol"
	tokenType_STRING     tokenType = "string"
	tokenType_NUMBER     tokenType = "number"
	tokenType_EOF        tokenType = "eof"
	tokenType_UNKNOWN    tokenType = "unknown"
)

type token struct {
	Type    tokenType
	Literal string
}

func (t token) String() string {
	switch t.Type {
	case tokenType_RESERVED:
		return fmt.Sprintf("<%s(%s)>", t.Type, t.Literal)
	case tokenType_IDENTIFIER:
		return fmt.Sprintf("<%s(%s)>", t.Type, t.Literal)
	case tokenType_SYMBOL:
		return fmt.Sprintf("<%s(%s)>", t.Type, t.Literal)
	case tokenType_STRING:
		return fmt.Sprintf("<%s(\"%s\")>", t.Type, t.Literal)
	case tokenType_NUMBER:
		return fmt.Sprintf("<%s(%s)>", t.Type, t.Literal)
	case tokenType_EOF:
		return fmt.Sprintf("<%s>", t.Type)
	case tokenType_UNKNOWN:
		return fmt.Sprintf("<%s(%s)>", t.Type, t.Literal)
	}
	return "unknown"
}

func (t token) toSQL() string {
	switch t.Type {
	case tokenType_RESERVED:
		return t.Literal
	case tokenType_IDENTIFIER:
		return t.Literal
	case tokenType_SYMBOL:
		return t.Literal
	case tokenType_STRING:
		return "'" + strings.ReplaceAll(t.Literal, "'", "''") + "'"
	case tokenType_NUMBER:
		return t.Literal
	case tokenType_EOF:
		return ""
	case tokenType_UNKNOWN:
		return t.Literal
	}
	return "unknown"
}

type lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *lexer {
	return &lexer{input: input, pos: 0}
}

func (l *lexer) NextToken() token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return token{Type: tokenType_EOF, Literal: ""}
	}

	symbols := []string{",", "=", "!=", "<", ">", "<=", ">=", "(", ")", "*", "?", ";"}
	for _, s := range symbols {
		if strings.HasPrefix(l.input[l.pos:], s) {
			l.pos += len(s)
			return token{Type: tokenType_SYMBOL, Literal: s}
		}
	}

	str := l.input[l.pos:]
	reserved := []string{"SELECT", "FROM", "AS", "UPDATE", "SET", "DELETE", "INSERT", "INTO", "VALUES", "WHERE", "AND", "IN", "LIKE", "GROUP BY", "ORDER BY", "ASC", "DESC", "LIMIT", "OFFSET"}
	for _, r := range reserved {
		if strings.HasPrefix(strings.ToUpper(str), r+" ") {
			l.pos += len(r) + 1
			return token{Type: tokenType_RESERVED, Literal: r}
		}
	}

	string_quotes := []byte{'\'', '"'}
	for _, q := range string_quotes {
		if s, ok := l.readString(q); ok {
			return token{Type: tokenType_STRING, Literal: s}
		}
	}

	ident_quotes := []byte{'`'}
	for _, q := range ident_quotes {
		if s, ok := l.readString(q); ok {
			return token{Type: tokenType_IDENTIFIER, Literal: s}
		}
	}

	ch := l.input[l.pos]
	if isLetter(ch) {
		start := l.pos
		for l.pos < len(l.input) && isLetter(l.input[l.pos]) {
			l.pos++
		}
		literal := l.input[start:l.pos]
		return token{Type: tokenType_IDENTIFIER, Literal: literal}
	}
	if isNumber(ch) {
		start := l.pos
		for l.pos < len(l.input) && isNumber(l.input[l.pos]) {
			l.pos++
		}
		literal := l.input[start:l.pos]
		return token{Type: tokenType_NUMBER, Literal: literal}
	}

	l.pos++
	return token{Type: tokenType_UNKNOWN, Literal: string(ch)}
}

func (l *lexer) skipWhitespace() {
	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t' || l.input[l.pos] == '\n') {
		l.pos++
	}
}

func (l *lexer) readString(quote byte) (string, bool) {
	if l.input[l.pos] != quote {
		return "", false
	}
	l.pos++
	start := l.pos
	for l.pos < len(l.input) && (l.input[l.pos] != quote || l.input[l.pos-1] == '\\') {
		l.pos++
	}
	result := l.input[start:l.pos]
	l.pos++
	return result, true
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isNumber(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
