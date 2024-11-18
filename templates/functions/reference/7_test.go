package main

import (
	"testing"
)

func TestBinarySearch(t *testing.T) {
	want := 2
	got := binarySearch([]int{1, 2, 3, 4, 5}, 3)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestBinarySearchNegative(t *testing.T) {
	want := -1
	got := binarySearch([]int{1, 2, 3, 4, 5}, 6)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestBinarySearchNotSorted(t *testing.T) {
	want := 1
	got := binarySearch([]int{1, 3, 2, 4, 5}, 3)
	if want == got {
		t.Fatalf("want %d, got %d, should be different", want, got)
	}
}
