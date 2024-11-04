package main

import "testing"

// FR1: App should print out 'This will be an issue tracker!
func TestGetMessage(t *testing.T) {
	want := "This will be an issue tracker!"
	got := getMessage()
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}
