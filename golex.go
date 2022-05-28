package golex

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// A lexer based on Rob Pikes 2011 talk for Google Technology User Group

// represents a token returned from the lexer
type Token struct {
	Typ tokenType
	Val string
}

// tokenType represents the type of tokens
type tokenType int

const (
		tokenError tokenType = iota // value is text of error

		tokenCloseBlock						// end of block, }}
		tokenDot									// the cursor, '.'
		tokenEOF									// END OF FILE
		tokenElse									// else keyword
		tokenEnd									// end keyword
		tokenHelper								// helper method
		tokenIdentifier						// identifier parameter
		tokenIf										// if keyword
		tokenNumber								// number
		tokenOpenBlock						// start of block, {{
		tokenRawString 						// raw quoted string
		tokenString								// quoted string
		tokenText									// plain text
		)

const eof = rune(tokenEOF)
const openBlock = "{{"
const closeBlock = "}}"

func (t Token) String() string {
	switch t.Typ {
	case tokenEOF:
		return "EOF"
	case tokenError:
		return t.Val
	}
	if len(t.Val) > 200 {
		return fmt.Sprintf("%.200q...", t.Val)
	}
	return fmt.Sprintf("%q", t.Val)
}

// represents the state of the lexer
// as a function that returns the next state
type stateFn func(*Lexer) stateFn

func (l *Lexer)Run() {
	for state:= lexText; state != nil; {
		state = state(l)
	}
	close(l.Tokens)
}

type Lexer struct {
	Name string
	Input string
	State stateFn
	start, current, width int
	Tokens chan Token
}

func New(name, input string) *Lexer {
	return &Lexer{
		Name: name,
		Input: input,
		State: lexText,
		start: 0,
		current: 0,
		Tokens: make(chan Token, 2),
	}
}

func (l *Lexer) NextToken() Token {
	for {
		select {
		case token := <-l.Tokens:
			return token
		default:
			l.State = l.State(l)
		}
	}
}

func (l *Lexer) emit(tt tokenType) {
	token := Token{tt, l.Input[l.start:l.current]}
	l.Tokens <- token

	l.start = l.current
}

func lexText(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.Input[l.current:], openBlock) {
			if l.current > l.start {
				l.emit(tokenText)
			}
			return lexOpenBlock
		}
		if l.next() == eof { break }
	}
	// Correctly reached EOF.
	if l.current > l.start {
		l.emit(tokenText)
	}
	l.emit(tokenEOF)
	return nil
}

func lexOpenBlock(l *Lexer) stateFn {
	l.current += len(openBlock)
	l.emit(tokenOpenBlock)
	return lexInsideAction
}

func lexCloseBlock(l *Lexer) stateFn {
	l.current += len(closeBlock)
	l.emit(tokenCloseBlock)
	return lexText
}

func lexInsideAction(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.Input[l.current:], closeBlock) {
			return lexCloseBlock
		}
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("unclosed action")
		case isSpace(r):
			l.ignore()
		case r == '+' || r == '-' || '0' <= r && r <= '9':
			l.backup()
			return lexNumber
		case isAlpha(r):
			l.backup()
			l.start = l.current
			return lexIdentifier
		}
	}
}

func lexIdentifier(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.Input[l.current:], closeBlock) {
			if l.current > l.start {
				l.emit(tokenIdentifier)
			}
			return lexCloseBlock
		}
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("unclosed action")
		case isSpace(r):
			l.ignore()
		}
	}
}

func lexNumber(l *Lexer) stateFn {
	l.accept("+-")
	digits := "0123456789"

	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)

	if l.accept("."){
		l.acceptRun(digits)
	}

	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	// imaginary number
	l.accept("i")

	if isAlpha(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.Input[l.start:l.current])
	}
	l.emit(tokenNumber)
	return lexInsideAction
}

// Lexer helpers
func (l *Lexer) next() (rune) {
	var res rune
	if l.current >= len(l.Input) {
		l.width = 0
		return eof
	}
	res, l.width = utf8.DecodeRuneInString(l.Input[l.current:])
	l.current += l.width
	return res
}

func (l *Lexer) ignore() {
	l.start = l.current
}

func (l *Lexer) backup(){
	l.current -= l.width
}

// Returns the next character without moving the lexer forward
func (l *Lexer) peek() rune {
	res := l.next()
	l.backup()
	return res
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {}
	l.backup()
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r)
}

// returns an error token and terminates the scan
// by passing nil pointer which will become the next state, terminating run loop
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.Tokens <- Token{
		tokenError,
		fmt.Sprintf(format, args...),
	}
	return nil
}
