package main

import "testing"

func TestGetMessage(t *testing.T) {
	want := "This will be a TODO list!"
	got := getMessage()
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}
