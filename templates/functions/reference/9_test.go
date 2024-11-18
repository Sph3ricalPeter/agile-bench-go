package main

import (
	"math"
	"reflect"
	"testing"
)

func TestDijkstra(t *testing.T) {
	want := map[string]int{"a": 0, "b": 1}
	got := dijkstra(map[string]map[string]int{
		"a": {"b": 1},
		"b": {"a": 1},
	}, "a")
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestDijkstraMissingMode(t *testing.T) {
	want := map[string]int{"c": 0, "a": math.MaxInt32}
	got := dijkstra(map[string]map[string]int{
		"a": {"b": 1},
	}, "c")
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
