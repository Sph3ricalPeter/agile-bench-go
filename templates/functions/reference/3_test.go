package main

import "testing"

func TestFib1(t *testing.T) {
	want := 5
	got := fib(5)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestFib2(t *testing.T) {
	want := 13
	got := fib(7)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestFib10(t *testing.T) {
	want := 55
	got := fib(10)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}
