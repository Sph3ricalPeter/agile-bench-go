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
	Score       int
}

// MustInitTemplate initializes the project template directory with the given project name
//
// new project template directory is created under templates/ directory, unless it already exists
// new project template will contain a project.yml template file, the init/ with empty main.go file and
// the reference/ folder with empty main.go files and 1_test.go file
func MustInitTemplate(projectDir string, projectType ProjectType) {
	err := os.Mkdir(fmt.Sprintf("templates/%s", projectDir), 0755)
	if err != nil {
		panic(fmt.Errorf("error creating project template directory: %w", err))
	}

	project := Project{
		Name:        "Sample Project",
		Description: "",
		Type:        projectType,
		Requirements: []Requirement{
			{
				Name:        "First requirement",
				Description: "First requirement description",
			},
		},
	}

	projectYml, err := yaml.Marshal(project)
	if err != nil {
		panic(fmt.Errorf("error marshalling project.yml: %w", err))
	}

	mustWriteFile(fmt.Sprintf("templates/%s/project.yml", projectDir), projectYml, 0644)

	mustMkdir(fmt.Sprintf("templates/%s/init", projectDir), 0755)
	mustWriteFile(fmt.Sprintf("templates/%s/init/main.go", projectDir), []byte("package main\n\nfunc main() {\n\n}\n"), 0644)

	mustMkdir(fmt.Sprintf("templates/%s/reference", projectDir), 0755)
	mustWriteFile(fmt.Sprintf("templates/%s/reference/main.go", projectDir), []byte("package main\n\nfunc main() {\n\n}\n"), 0644)
	mustWriteFile(fmt.Sprintf("templates/%s/reference/1_test.go", projectDir), []byte("package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {\n\n}\n"), 0644)

	fmt.Printf("Project template %s initialized\n", projectDir)
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

// TakeCodebaseSnapshot copies the app/ directory to a new directory with the given name inside the out/ directory
func TakeCodebaseSnapshot(filename string) error {
	fmt.Printf("Taking codebase snapshot %s\n", filename)
	common.RunCommand("ls -la app/")
	dir := fmt.Sprintf("out/%s", filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating snapshots directory: %w", err)
	}

	err = common.RunBashCommand(fmt.Sprintf("cp -r app/ %s", dir))
	if err != nil {
		return fmt.Errorf("error copying app directory to snapshots: %w", err)
	}

	return nil
}

func mustWriteFile(name string, data []byte, perm os.FileMode) {
	err := os.WriteFile(name, data, perm)
	if err != nil {
		panic(fmt.Errorf("error writing file: %w", err))
	}
}

func mustMkdir(name string, perm os.FileMode) {
	err := os.Mkdir(name, perm)
	if err != nil {
		panic(fmt.Errorf("error creating directory: %w", err))
	}
}
