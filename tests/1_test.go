// 1_test.go
package main

import "testing"

// FR1: App should print out 'This will be a TODO list!'
func TestGetMessage_1(t *testing.T) {
	want := "This will be a TODO list!"
	got := getMessage()
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}
