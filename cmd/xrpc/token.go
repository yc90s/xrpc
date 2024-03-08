package main

import "fmt"

const (
	LCURLY = "{"
	RCURLY = "}"
	LPAREN = "("
	RPAREN = ")"
	COMMA  = ","

	SERVICE = "service"
	PACKAGE = "package"
	IMPORT  = "import"
	GO      = "go"

	ID   = "ID"
	PATH = "PATH"
	EOF  = "EOF"
)

var RESERVE_KEYWORDS = map[string]string{
	SERVICE: SERVICE,
	PACKAGE: PACKAGE,
	IMPORT:  IMPORT,
	GO:      GO,
}

var SINGLE_CHAR_TOKENS = map[string]string{
	LCURLY: LCURLY,
	RCURLY: RCURLY,
	LPAREN: LPAREN,
	RPAREN: RPAREN,
	COMMA:  COMMA,
}

type Token struct {
	tp     string // token type
	value  string
	lineno int
	column int
}

func NewToken(tp, value string, lineno, column int) *Token {
	return &Token{
		tp:     tp,
		value:  value,
		lineno: lineno,
		column: column,
	}
}

func (t *Token) String() string {
	return fmt.Sprintf("Token('%s', '%s') in line:%d, column:%d", t.tp, t.value, t.lineno, t.column)
}
