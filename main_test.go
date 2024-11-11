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
	project.MustInitProject("simple-todo")
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

func TestParseWriteResponse(t *testing.T) {
	exampleBlock := "```go\n// start of lib.go\npackage main\n\nfunc add(a, b int) int {\n\treturn a + b\n}\n// end of lib.go\n// start of calc.go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tres := add(2, 3)\n\tfmt.Println(res)\n}\n// end of calc.go\n```\n"
	_, err := internal.ParseWriteResponse([]byte(exampleBlock))
	if err != nil {
		t.Fatalf("error parsing write response: %s", err.Error())
	}
}
