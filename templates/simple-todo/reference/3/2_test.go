package main

import "testing"

func TestAddTodo(t *testing.T) {
	storage := Storage{}

	want := "Buy milk"
	storage.addTodo(want)

	got := storage.getTodos()[0]
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}
