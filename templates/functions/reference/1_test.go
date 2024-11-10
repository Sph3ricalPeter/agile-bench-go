package main

import "testing"

func TestAdd(t *testing.T) {
	want := 5
	got := add(2, 3)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}
