package main

import (
	"reflect"
	"testing"
)

func TestCustomSplit1(t *testing.T) {
	want := []string{"a", "b", "c"}
	got := splitIterative("a b c", " ")
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
