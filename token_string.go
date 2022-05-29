package golex

import "fmt"

func (t Token) String() string {
	switch t.Typ {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return t.Val
	}
	if len(t.Val) > 200 {
		return fmt.Sprintf("%.200q...", t.Val)
	}
	return fmt.Sprintf("%s", t.Val)
}