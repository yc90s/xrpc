package main

import (
	"fmt"
	"testing"
)

func Test_isalnum(t *testing.T) {
	testData := []map[byte]bool{
		{'a': true},
		{'1': true},
		{' ': false},
		{'$': false},
	}

	for _, data := range testData {
		for k, v := range data {
			if isalnum(k) != v {
				t.Errorf("isalnum(%c) != %v", k, v)
			}
		}
	}
}

func Test_iswhitespace(t *testing.T) {
	testData := []map[byte]bool{
		{'a': false},
		{'1': false},
		{' ': true},
		{'\n': true},
		{'\t': true},
		{'\r': true},
		{'\v': true},
	}

	for _, data := range testData {
		for k, v := range data {
			if iswhitespace(k) != v {
				t.Errorf("iswhitespace(%c) != %v", k, v)
			}
		}
	}
}

func Test_Lexer_advance(t *testing.T) {
	data := []byte("hello \n\r world")
	lexer := NewLexer(data)
	count := 0

	for lexer.current_char != None {
		if lexer.current_char != data[count] {
			t.Errorf("lexer.current_char != data[count]")
		}
		lexer.advance()
		count++
	}

	if count != len(data) {
		t.Errorf("count != len(data)")
	}
}

func checkTokenEqual(l, r *Token) bool {
	if l.tp != r.tp {
		fmt.Println(l.tp, r.tp)
		return false
	}
	if l.value != r.value {
		fmt.Println(l.value, r.value)
		return false
	}
	if l.lineno != r.lineno {
		fmt.Println(l.lineno, r.lineno)
		return false
	}
	if l.column != r.column {
		fmt.Println(l.column, r.column)
		return false
	}
	return l.tp == r.tp && l.value == r.value && l.lineno == r.lineno && l.column == r.column
}

func Test_Lexer_get_next_token(t *testing.T) {
	data := []byte("{{}}(),,,service package,\n\timport, \n\"service package\", servicepackageimport  \t,,,")
	verifyTokens := []Token{
		{LCURLY, "{", 1, 1},
		{LCURLY, "{", 1, 2},
		{RCURLY, "}", 1, 3},
		{RCURLY, "}", 1, 4},
		{LPAREN, "(", 1, 5},
		{RPAREN, ")", 1, 6},
		{COMMA, ",", 1, 7},
		{COMMA, ",", 1, 8},
		{COMMA, ",", 1, 9},
		{SERVICE, "service", 1, 10},
		{PACKAGE, "package", 1, 18},
		{COMMA, ",", 1, 25},
		{IMPORT, "import", 2, 2},
		{COMMA, ",", 2, 8},
		{PATH, "service package", 3, 1},
		{COMMA, ",", 3, 18},
		{ID, "servicepackageimport", 3, 20},
		{COMMA, ",", 3, 43},
		{COMMA, ",", 3, 44},
		{COMMA, ",", 3, 45},
	}
	lexer := NewLexer(data)

	count := 0
	for lexer.current_char != None {
		token := lexer.get_next_token()
		if !checkTokenEqual(token, &verifyTokens[count]) {
			fmt.Println(token, "v: ", verifyTokens[count])
			t.Errorf("token != &verifyTokens[count]")
		}
		count++
	}

	if count != len(verifyTokens) {
		t.Errorf("count != len(verifyTokens)")
	}
}
