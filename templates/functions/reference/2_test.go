package main

import "testing"

func TestSubtract(t *testing.T) {
	want := 2
	got := subtract(3, 1)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}
