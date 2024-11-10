package main

import (
	"testing"

	"github.com/Sph3ricalPeter/frbench/internal"
	"github.com/Sph3ricalPeter/frbench/internal/project"
)

func TestLoadCodebase(t *testing.T) {
	codebase, err := project.LoadCodebase()
	if err != nil {
		t.Fatalf("error loading codebase: %s", err.Error())
	}
	if codebase == nil {
		t.Fatalf("codebase is nil")
	}
}

func TestInitProject(t *testing.T) {
	err := project.InitProject("simple-todo")
	if err != nil {
		t.Fatalf("error initializing project: %s", err.Error())
	}
}

func TestPreparePrompt(t *testing.T) {
	_, err := internal.PreparePatchPrompt(project.ProjectInfo{
		Project: project.Project{
			Requirements: []project.Requirement{
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
