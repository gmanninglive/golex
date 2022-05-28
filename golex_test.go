package golex

import (
	"testing"
)

const testString = "<div>{{name}}</div>"

func TestLex(t *testing.T) {

  l := New("test", testString)
	var received []Token
  for l.State != nil {
    received = append(received, l.NextToken())
  }

	t.Log("length", len(received))

	if len(received) != 5 {
		t.Errorf("Expected 5 tokens, got %d", len(received))
	}
}