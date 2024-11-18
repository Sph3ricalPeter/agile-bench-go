package main

import (
	"reflect"
	"testing"
)

func TestMergeSort(t *testing.T) {
	want := []int{1, 2, 3, 4, 5}
	got := mergeSort([]int{5, 4, 3, 2, 1})
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestMergeSortRandom(t *testing.T) {
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := mergeSort([]int{3, 6, 1, 4, 5, 2, 7, 9, 8})
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
