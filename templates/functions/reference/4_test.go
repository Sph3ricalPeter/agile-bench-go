package main

import (
	"reflect"
	"testing"
)

func TestSplit1(t *testing.T) {
	want := []string{"a", "b", "c"}
	got := split("a b c", " ")
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSplit2(t *testing.T) {
	want := []string{"first part", "second part"}
	got := split("first part, second part", ", ")
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
