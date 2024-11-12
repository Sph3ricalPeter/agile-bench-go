package main

import (
	"testing"

	"github.com/google/uuid"
)

func TestAddTodo(t *testing.T) {
	storage := newTestStorage()

	want := "Buy milk"
	storage.addTodo(want)

	got := storage.getTodos()[1].text
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestRemoveTodo(t *testing.T) {
	storage := newTestStorage()

	want := 0
	storage.removeTodo(storage.getTodos()[0].id)

	got := len(storage.getTodos())
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestMarkTodoAsDone(t *testing.T) {
	storage := newTestStorage()

	want := true
	storage.markTodoAsDone(storage.getTodos()[0].id)

	got := storage.getTodos()[0].isDone
	if want != got {
		t.Fatalf("want %t, got %t", want, got)
	}
}

func newTestStorage() Storage {
	return Storage{
		todos: []Todo{
			{id: uuid.NewString(), text: "Do laundry", isDone: false},
		},
	}
}
