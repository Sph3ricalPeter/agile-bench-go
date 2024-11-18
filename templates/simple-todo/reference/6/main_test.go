package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestGetMessage(t *testing.T) {
	want := "Welcome to a TODO list!"
	got := getMessage()
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

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

func TestMarkTodo(t *testing.T) {
	storage := newTestStorage()

	want := Done
	storage.setTodoStatus(storage.getTodos()[0].id, Done)

	got := storage.getTodos()[0].status
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestGetTodosByStatus(t *testing.T) {
	storage := newTestStorage()

	want := "Do laundry"
	got := storage.getTodosByStatus(Done)[0]
	if want != got.text {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestWriteTodos(t *testing.T) {
	s := Storage{todos: []Todo{
		{id: uuid.NewString(), text: "Buy milk", status: Open},
		{id: uuid.NewString(), text: "Do laundry", status: Done},
	}}
	want := fmt.Sprintf("%s: Buy milk (open)\n%s: Do laundry (done)\n", s.todos[0].id, s.todos[1].id)

	var b bytes.Buffer
	s.writeTodos(&b)
	got := b.String()
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func newTestStorage() Storage {
	return Storage{
		todos: []Todo{
			{id: uuid.NewString(), text: "Do laundry", status: Done},
		},
	}
}
