package main

import "testing"

func TestLoadCodebase(t *testing.T) {
	codebase, err := loadCodebase()
	if err != nil {
		t.Fatalf("error loading codebase: %s", err.Error())
	}
	if codebase == nil {
		t.Fatalf("codebase is nil")
	}
}

func TestInitProject(t *testing.T) {
	err := initProject("simple-todo")
	if err != nil {
		t.Fatalf("error initializing project: %s", err.Error())
	}
}

func TestPreparePrompt(t *testing.T) {
	_, err := preparePrompt(ProjectInfo{
		Project: Project{
			Requirements: []Requirement{
				{
					Description: "App should print out a message 'This will be a TODO app!' when started.",
					Tests:       []string{"1_test.go"},
				},
			},
		},
	}, 0)
	if err != nil {
		t.Fatalf("error preparing prompt: %s", err.Error())
	}
}
