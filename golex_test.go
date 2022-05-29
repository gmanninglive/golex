package golex

import (
	"os"
	"testing"
)

const testString = "<div>{{name}}</div>"

const (
	TokenText TokenType = iota
	TokenOpenBlock
	TokenCloseBlock
	TokenNewLine
	TokenCharO
)

const (
	openBlock  = "{{"
	closeBlock = "}}"
	newLine    = "\n"
)

func mockTextStateFn(l *Lexer) StateFn {
	for {
		if l.nextHasPrefix(openBlock) {
			if l.current > l.start {
				l.Emit(TokenText)
			}
			return mockOpenBlockStateFn
		}

		if l.nextHasPrefix(closeBlock) {
			if l.current > l.start {
				l.Emit(TokenText)
			}
			return mockCloseBlockStateFn
		}

		if l.nextHasPrefix(newLine) {
			if l.current > l.start {
				l.Emit(TokenText)
			}
			return mockNewLineStateFn
		}

		if l.nextHasPrefix("o") {
			if l.current > l.start {
				l.Emit(TokenText)
			}
			return mockCharOStateFn
		}

		if l.Next() == eof {
			break
		}
	}
	if l.current > l.start {
		l.Emit(TokenText)
	}
	l.Emit(TokenEOF)
	return nil
}

func mockOpenBlockStateFn(l *Lexer) StateFn {
	l.current += len(openBlock)
	l.Emit(TokenOpenBlock)
	return mockTextStateFn
}

func mockCloseBlockStateFn(l *Lexer) StateFn {
	l.current += len(closeBlock)
	l.Emit(TokenCloseBlock)
	return mockTextStateFn
}

func mockNewLineStateFn(l *Lexer) StateFn {
	l.current += len(newLine)
	l.Emit(TokenNewLine)
	return mockTextStateFn
}

func mockCharOStateFn(l *Lexer) StateFn {
	l.current += len("o")
	l.Emit(TokenCharO)
	return mockTextStateFn
}

func TestLex(t *testing.T) {

	t.Run("Moves through string, it returns all tokens", func(t *testing.T) {
		l := New("test", testString, mockTextStateFn)

		var received []Token

		for {
			token, done := l.NextToken()
			if done { break }
			received = append(received, token)
		}

		t.Log("tokens:", received)

		if len(received) != 5 {
			t.Errorf("Expected 5 tokens, got %d", len(received))
		}

		var out string
		for _, tok := range received {
			// t.Logf("Each Token string representation %s\n", tok.String())
			out += string(tok.String())
		}

		if out != testString {
			t.Errorf("Value corrupted during lexing,\nexpected: %s\n, got: %s\n", testString, out)
		}
	},
	)

	f, err := os.ReadFile("./test/fixtures/plaintext")
	if err != nil {
		panic(err)
	}

	t.Run("Using RunSync() Method", func(t *testing.T) {
		l := New("test", string(f), mockTextStateFn)
		l.RunSync()

		var received []Token
		for {
			tok, done := l.Listen()
			if done {
				t.Logf("Total Tokens: %o", len(received))
				return
			}
			received = append(received, tok)
		}
	})

	t.Run("Using RunAsync() Method", func(t *testing.T) {
		l := New("test", string(f), mockTextStateFn)
		l.RunAsync()

		var received []Token
		for {
			tok, done := l.Listen()
			if done {
				t.Logf("Total Tokens: %o", len(received))
				return
			}
			received = append(received, tok)
		}
	})
}
