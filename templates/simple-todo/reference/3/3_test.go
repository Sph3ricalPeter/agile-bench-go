package main

import "testing"

func TestRemoveTodo(t *testing.T) {
	storage := Storage{
		todos: []string{"Buy milk", "Do laundry"},
	}

	want := "Buy milk"
	storage.removeTodo(storage.todos[1])

	got := storage.getTodos()[0]
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}
