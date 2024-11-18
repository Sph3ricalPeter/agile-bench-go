package main

import "testing"

func TestFibConc1(t *testing.T) {
	want := 5
	got := fibConc(5)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestFibConc7(t *testing.T) {
	want := 13
	got := fib(7)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestFibConc10(t *testing.T) {
	want := 55
	got := fib(10)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}
