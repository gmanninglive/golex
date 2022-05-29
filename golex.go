package golex

// A lexer based on Rob Pikes 2011 talk for Google Technology User Group

// Usage
//
// Declare constant of available tokens as follow:
// const (
// 	TokenText       TokenType = iota
// 	TokenOpenBlock
// 	TokenCloseBlock
//  ...
// )
//
// Create a state function to handle each token type
// The state function must return a state function to move on or nil to end execution
//
// Example:
// // state function to move through a block of text till it reaches the EOF
// // emitting a text token and a EOF token before returning nil to end the lexer execution
// func textStateFn(l *Lexer) StateFn {
// 	for {
// 		if l.Next() == eof {
// 			break
// 		}
// 	}
// 	if l.current > l.start {
// 		l.Emit(TokenText)
// 	}
// 	l.Emit(TokenEOF)
// 	return nil
// }
//
// Initialise a lexer with New() passing in name, input, and intial state
//
// func main() {
//
// l := golex.New("txt", s, textStateFn)
//
// }
//

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// represents a token returned from the lexer
type Token struct {
	Typ TokenType
	Val string
}

// tokenType represents the type of tokens
type TokenType int

const (
	TokenEOF   TokenType = -2 // END OF FILE
	TokenError TokenType = -1 // Value contains error text
)

const eof = rune(TokenEOF)

// Represents the state of the lexer
// As a function that returns a function
type StateFn func(*Lexer) StateFn

type Lexer struct {
	Name                  string
	Input                 string
	State                 StateFn
	InitialState          StateFn
	start, current, width int
	Tokens                chan Token
}

func New(name, input string, initialState StateFn) *Lexer {
	return &Lexer{
		Name:         name,
		Input:        input,
		State:        initialState,
		InitialState: initialState,
		start:        0,
		current:      0,
		Tokens:       make(chan Token, 2),
	}
}

func (l *Lexer) RunSync() {
	l.Tokens = make(chan Token, len(l.Input)/2)
	l.run()
}

func (l *Lexer) RunAsync() {
	l.Tokens = make(chan Token, len(l.Input)/2)
	go l.run()
}

// Private run method
func (l *Lexer) run() {
	for state := l.InitialState; state != nil; {
		state = state(l)
	}
	close(l.Tokens)
}

// Listen returns the most recent token received from the channel
// And a boolean value for if the lexer has finished scanning
func (l *Lexer) Listen() (t Token, done bool) {
	select {
	case tok := <-l.Tokens:
		if tok.Typ == TokenEOF {
			return tok, true
		}
		return tok, false
	}
}

// Sync method to move through the input and return tokens
func (l *Lexer) NextToken() (Token, bool) {
	for {
		select {
		case token := <-l.Tokens:
			if token.Typ == TokenEOF {
				return token, true 
			} else {
				return token, false
			}
		default:
			l.State = l.State(l)
		}
	}
}

// Sends token to the Tokens channel and moves starting position to current position
func (l *Lexer) Emit(tt TokenType) {
	token := Token{tt, l.Input[l.start:l.current]}
	l.Tokens <- token

	l.start = l.current
}

// Lexer helpers
func (l *Lexer) Next() rune {
	var res rune
	if l.current >= len(l.Input) {
		l.width = 0
		return eof
	}
	res, l.width = utf8.DecodeRuneInString(l.Input[l.current:])
	l.current += l.width
	return res
}

func (l *Lexer) Ignore() {
	l.start = l.current
}

func (l *Lexer) Backup() {
	l.current -= l.width
}

// Returns the next character without moving the lexer forward
func (l *Lexer) Peek() rune {
	res := l.Next()
	l.Backup()
	return res
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.Next()) >= 0 {
		return true
	}
	l.Backup()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.Next()) >= 0 {
	}
	l.Backup()
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r)
}

func (l *Lexer) nextHasPrefix(prefix string) bool {
	next := l.Input[l.current:]
	return strings.HasPrefix(next, prefix)
}

// Returns an error token and terminates the scan
// By passing nil pointer which will become the next state, terminating run loop
func (l *Lexer) errorf(format string, args ...interface{}) StateFn {
	l.Tokens <- Token{
		TokenError,
		fmt.Sprintf(format, args...),
	}
	return nil
}
