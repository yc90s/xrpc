package main

import (
	"unicode"
)

type Lexer struct {
	text []byte

	pos          int
	lineno       int
	column       int
	current_char byte
}

const None = 0

func NewLexer(text []byte) *Lexer {
	lexer := &Lexer{
		text:   text,
		pos:    0,
		lineno: 1,
		column: 1,
	}
	lexer.current_char = lexer.text[lexer.pos]
	return lexer
}

func isalnum(b byte) bool {
	return unicode.IsNumber(rune(b)) || unicode.IsLetter(rune(b))
}

func iswhitespace(b byte) bool {
	return unicode.IsSpace(rune(b))
}

func (l *Lexer) advance() {
	if l.current_char == '\n' {
		l.lineno++
		l.column = 0
	}

	l.pos++
	if l.pos > len(l.text)-1 {
		l.current_char = None
	} else {
		l.current_char = l.text[l.pos]
		l.column++
	}
}

func (l *Lexer) get_next_token() *Token {
	for l.current_char != None {
		if iswhitespace(l.current_char) {
			l.skip_whitespace()
			continue
		}

		if l.current_char == '"' {
			return l.path()
		}

		if val, ok := SINGLE_CHAR_TOKENS[string(l.current_char)]; ok {
			token := NewToken(val, val, l.lineno, l.column)
			l.advance()
			return token
		}

		return l.id()
	}

	return NewToken(EOF, "", l.lineno, l.column)
}

func (l *Lexer) id() *Token {
	var value string

	lineno := l.lineno
	column := l.column

	if l.current_char == '*' {
		value += string(l.current_char)
		l.advance()
	}

	for l.current_char != None && (isalnum(l.current_char) || l.current_char == '_' || l.current_char == '.') {
		value += string(l.current_char)
		l.advance()
	}

	if val, ok := RESERVE_KEYWORDS[value]; ok {
		return NewToken(val, value, lineno, column)
	}

	if len(value) == 0 {
		return NewToken(EOF, "", lineno, column)
	}

	return NewToken(ID, value, lineno, column)
}

func (l *Lexer) path() *Token {
	var value string

	lineno := l.lineno
	column := l.column

	l.advance()
	for l.current_char != '"' {
		value += string(l.current_char)
		l.advance()
	}

	l.advance()

	return NewToken(PATH, value, lineno, column)
}

func (l *Lexer) skip_whitespace() {
	for iswhitespace(l.current_char) {
		l.advance()
	}
}
