package main

import (
	"testing"
)

func TestReverseString(t *testing.T) {
	want := "olleH"
	got := reverse("Hello")
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestReverseStringSpaces(t *testing.T) {
	want := "evol em eviG"
	got := reverse("Give me love")
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestReverseStringEmpty(t *testing.T) {
	want := ""
	got := reverse("")
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}
