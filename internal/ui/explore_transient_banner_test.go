package ui

import "testing"

func TestCombineTransientFeedback_order(t *testing.T) {
	s := combineTransientFeedback("r", "rec", "p")
	if s != "p · rec · r" {
		t.Fatalf("got %q", s)
	}
}

func TestCombineTransientFeedback_empty(t *testing.T) {
	if combineTransientFeedback("", "", "") != "" {
		t.Fatal("expected empty")
	}
}
