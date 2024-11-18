package main

import (
	"bytes"
	"fmt"
	"os/exec"
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
		t.Fatalf("want '%s', got '%s'", want, got)
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

func TestCLIUsage(t *testing.T) {
	want := `1: Laundry (open)
`
	got := runCmdHelper(t, []string{"go", "run", "main.go"}, `list
add Laundry
list
remove 1
list
exit
`)
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func runCmdHelper(t *testing.T, args []string, in string) string {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	out := bytes.NewBuffer(nil)
	cmd.Stdout = out
	cmd.Stdin = bytes.NewBufferString(in)

	err := cmd.Run()
	if err != nil {
		t.Fatalf("cmd failed: %s", err.Error())
	}

	return out.String()
}

func newTestStorage() Storage {
	return Storage{
		idGen: &SeqIdGenerator{},
		todos: []Todo{
			{id: uuid.NewString(), text: "Do laundry", status: Done},
		},
	}
}
