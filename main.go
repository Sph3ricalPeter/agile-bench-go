package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Sph3ricalPeter/frbench/common"
	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/external/anth"
	"gopkg.in/yaml.v3"
)

type ModelInstance struct {
	con   external.Connector
	score int
}

type ProjectInfo struct {
	Project Project
}

type Project struct {
	Name         string
	Description  string
	Requirements []Requirement
}

type Requirement struct {
	Description string
	Tests       []string
}

func main() {
	// read cache flag from args
	useCache := flag.Bool("c", false, "use cache")
	flag.Parse()

	models := []ModelInstance{
		// {
		// 	con:   google.NewGoogleConnector(google.Gemini15Flash8B),
		// 	score: 0,
		// },
		{
			con:   anth.NewAnthConnector(anth.Claude3Haiku),
			score: 0,
		},
	}

	// TODO: somehow do this dynamically based on the provided test files
	// so probably load files based on regex fname matching and it over them
	testCount := 2

	// 1. copy the initial codebase for the project
	project := "functions"
	common.CheckErr(initProject(project))

	// 2. read project.yml and load the project info
	projectYml, err := os.ReadFile(fmt.Sprintf("templates/%s/project.yml", project))
	common.CheckErr(err)
	var projectInfo ProjectInfo
	common.CheckErr(yaml.Unmarshal(projectYml, &projectInfo))

	appliedPatches := map[int][]byte{}
	invalidPatches := map[int]*string{}
	for _, model := range models {
		for i := 1; i < testCount+1; i++ {
			fmt.Printf("Running task #%d on model %s ...\n", i, model.con.GetModelName())

			// move test files before prompt is created with codebase inside
			err := runCommand(fmt.Sprintf("cp templates/%s/reference/%d_test.go app/", project, i))
			if err != nil {
				fmt.Printf("error copying test file: %s\n", err)
				break
			}

			err = verifyPatch()
			if err == nil {
				panic("test passed before patching")
			}

			fmt.Println("Sending prompt ...")
			promptBytes, err := preparePrompt(projectInfo, i-1)
			if err != nil {
				fmt.Printf("error reading test file: %s\n", err.Error())
				break
			}
			result, err := model.con.SendPrompt(external.SendPromptData{
				Role:     external.RoleUser,
				Prompt:   promptBytes,
				UseCache: *useCache,
				Number:   i,
			})
			if err != nil {
				fmt.Printf("error sending prompt: %s\n", err.Error())
				break
			}
			fmt.Println("OK.")
			if result.CacheKey == nil {
				fmt.Printf("Used %d input / %d output tokens.\n", result.Usage.InputTokens, result.Usage.OutputTokens)
			}

			updatedPatch, err := doPatch([]byte(result.Content), i)
			if err != nil {
				fmt.Printf("error doing patch: %s\n", err.Error())
				invalidPatches[i] = result.CacheKey
				break
			}
			appliedPatches[i] = updatedPatch

			err = verifyPatch()
			if err != nil {
				fmt.Println("Patch BAD! ❌")
				invalidPatches[i] = result.CacheKey
				break
			} else {
				fmt.Println("Patch OK! ✅")
				model.score++
			}

			// wait for input to revert patches
			fmt.Println("Press ENTER to do next patch or revert if it's the last one ...")
			_, _ = os.Stdin.Read(make([]byte, 1))
		}

		fmt.Printf("All Done! %s scored: %d/%d\n", model.con.GetModelName(), model.score, testCount)

		// for i := len(appliedPatches) - 1; i >= 0; i-- {
		// 	fmt.Printf("Reverting patch #%d ...\n", i+1)
		// 	common.CheckErr(runCommand(fmt.Sprintf("rm app/%d_test.go", i+1)))
		// 	common.CheckErr(revertPatch(appliedPatches[i]))
		// }

		// prompt is the key for the cache
		for i, cacheKey := range invalidPatches {
			if cacheKey == nil {
				continue
			}
			fmt.Printf("Removing cache for invalid patch #%d ...\n", i)
			common.CheckErr(model.con.InvalidateCachedPrompt(*cacheKey))
		}
	}
}

func preparePrompt(projectInfo ProjectInfo, i int) ([]byte, error) {
	requirement := projectInfo.Project.Requirements[i]
	codebase, err := loadCodebase()
	if err != nil {
		return nil, fmt.Errorf("error loading codebase: %w", err)
	}
	prompt := `You will be provided with a full codebase inside the app/ directory and an issue statement explaining what needs to be changed in that codebase. The changes made need to make all provided tests pass but you can't change any tests.
<issue>
%s
</issue>
<codebase>
%s
</codebase>
Here is an example of a patch file. It consists of changes to files in the codebase. It specifies the file names, the line numbers of each change, and the removed and added lines. A single patch file can contain changes to multiple files.
<patch>
--- a/app/file.go
+++ b/app/file.go
@@ -1,8 +1,8 @@
 package main
 
 func Euclidean(a, b int) int {
-	for b != 0 {
-		a, b = b, a/b
+	if b == 0 {
+		return a
 	}
-	return a
+	return Euclidean(b, a/b)
 }

</patch>
I need you to implement the required changes by generating a single patch file that can be applied directly using git apply. Please respond with a single patch file in the format shown above. Don't add any additional text or comments only the patch file contents.
Respond below:
`

	ret := []byte(fmt.Sprintf(prompt, requirement.Description, codebase))
	_ = os.WriteFile("data/prompt.txt", ret, 0644)

	return ret, nil
}

// initProject initializes the project directory with the given project template
// and removes any existing files in the app/ directory
func initProject(project string) error {
	err := os.RemoveAll("app/")
	if err != nil {
		return fmt.Errorf("error removing app directory: %w", err)
	}

	err = os.Mkdir("app/", 0755)
	if err != nil {
		return fmt.Errorf("error creating app directory: %w", err)
	}

	projectTemplateDir := fmt.Sprintf("templates/%s/init", project)
	err = runBashCommand(fmt.Sprintf("cp -r %s/* app/", projectTemplateDir))
	if err != nil {
		return fmt.Errorf("error copying project template: %w", err)
	}

	return nil
}

// loads all codebase files into a single string, separated by [start of file.extension] and [end of file.extension]
func loadCodebase() ([]byte, error) {
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

func doPatch(patch []byte, i int) ([]byte, error) {
	patchBytes, err := common.WritePatchFile(patch, fmt.Sprintf("data/patch-%d.patch", i))
	if err != nil {
		return nil, fmt.Errorf("error writing patch file: %w", err)
	}
	// err = applyPatch(patchBytes)
	err = runCommand(fmt.Sprintf("git apply data/patch-%d.patch", i))
	if err != nil {
		return nil, fmt.Errorf("error applying patch: %w", err)
	}
	return patchBytes, nil
}

func verifyPatch() error {
	err := runTests()
	if err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}
	return nil
}

func applyPatch(patch []byte) error {
	return runCommandWithInput("patch -ruN -d app/", bytes.NewReader(patch))
}

// func revertPatch(patch []byte) error {
// 	return runCommandWithInput("patch -ruN --reverse -d app/", bytes.NewReader(patch))
// }

func runTests() error {
	return runCommand("go test -v ./app/")
}

func runCommand(cmd string) error {
	return runCommandWithInput(cmd, os.Stdin)
}

func runCommandWithInput(cmd string, input io.Reader) error {
	parts := strings.Split(cmd, " ")

	c := exec.Command(parts[0], parts[1:]...)
	c.Stdin = input
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	fmt.Println(c.String())

	return c.Run()
}

// runBashCommand runs a bash command, which allows the expansion of globs (*) and other bash features
func runBashCommand(cmd string) error {
	err := exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		return fmt.Errorf("error running bash command: %w", err)
	}
	return nil
}
