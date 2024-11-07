// 3_test.go
package main

import "testing"

// FR3: App should use in-memory storage and support adding TODOs
func TestAddTodo_3(t *testing.T) {
	storage := Storage{}

	want := "Buy milk"
	storage.addTodo(want)

	got := storage.getTodos()[0]
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}
