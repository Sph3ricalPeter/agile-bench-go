package project

import (
	"fmt"
	"os"

	"github.com/Sph3ricalPeter/frbench/internal/common"
	"gopkg.in/yaml.v3"
)

type ProjectInfo struct {
	Dir     string
	Project Project
}

type ProjectType string

const (
	ProjectTypeSingle      ProjectType = "single"      // single main.go file project, all tests should always pass
	ProjectTypeCheckpoints ProjectType = "checkpoints" // project with possible breaking changes, requiring checkpoints with different tests
)

type Project struct {
	Name         string
	Description  string
	Type         ProjectType
	Requirements []Requirement
}

type Requirement struct {
	Name        string
	Description string
	Attachments []string
}

func MustLoadFromYaml(projectDir string) ProjectInfo {
	projectYml, err := os.ReadFile(fmt.Sprintf("templates/%s/project.yml", projectDir))
	if err != nil {
		panic(fmt.Errorf("error reading project.yml: %w", err))
	}
	var projectInfo ProjectInfo
	err = yaml.Unmarshal(projectYml, &projectInfo)
	if err != nil {
		panic(fmt.Errorf("error unmarshalling project.yml: %w", err))
	}
	projectInfo.Dir = projectDir
	return projectInfo
}

// initProject initializes the project directory with the given project template
// and removes any existing files in the app/ directory
func MustInitProject(project string) {
	err := os.RemoveAll("app/")
	if err != nil {
		panic(fmt.Errorf("error removing app directory: %w", err))
	}

	err = os.Mkdir("app/", 0755)
	if err != nil {
		panic(fmt.Errorf("error creating app directory: %w", err))
	}

	projectTemplateDir := fmt.Sprintf("templates/%s/init", project)
	err = common.RunBashCommand(fmt.Sprintf("cp -r %s/* app/", projectTemplateDir))
	if err != nil {
		panic(fmt.Errorf("error copying project template: %w", err))
	}
}

// loads all codebase files into a single string, separated by [start of file.extension] and [end of file.extension]
func LoadCodebase() ([]byte, error) {
	ret := []byte{}

	// load all files under app/ directory
	files, err := os.ReadDir("app/")
	if err != nil {
		return nil, fmt.Errorf("error reading app directory: %w", err)
	}
	fmt.Printf("Loaded files from app/ codebase: %v\n", files)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileBytes, err := os.ReadFile(fmt.Sprintf("app/%s", file.Name()))
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}

		// a line before and after each file [start of file.extension], [end of file.extension]
		ret = append(ret, []byte(fmt.Sprintf("// start of %s\n", file.Name()))...)
		ret = append(ret, fileBytes...)
		ret = append(ret, []byte(fmt.Sprintf("// end of %s\n", file.Name()))...)
	}

	// FIXME: testing only
	_ = os.WriteFile("data/codebase.txt", ret, 0644)

	return ret, nil
}
