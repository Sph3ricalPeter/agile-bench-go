package main

import "testing"

func TestGetMessage(t *testing.T) {
	want := "Welcome to a TODO list!"
	got := getMessage()
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}
