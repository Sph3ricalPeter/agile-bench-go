// 2_test.go
package main

import "testing"

// FR2: App should print a list of 5 TODOs
func TestGetTodos_2(t *testing.T) {
	want := 5
	got := len(getTodos())
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}
