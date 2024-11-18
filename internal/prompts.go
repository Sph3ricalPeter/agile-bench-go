package internal

import (
	"encoding/base64"
	"fmt"
	"os"
	"regexp"

	"github.com/Sph3ricalPeter/frbench/internal/project"
)

const (
	SystemPrompt = `You are a software engineer in an agile team tasked to create an issue-tracker application in Golang.
You will receive one functional requirement (FR) for what you need to implement in each prompt and are expected to provide all required code and file changes as a response in the form of a patch file.
Along with the FR, you will receive the contents of a test file that will verify the changes you made. All previous tests should also pass.
Assume the working directory is app/ and it starts empty.
Assume your changes from previous prompts are already applied.
Your response should only include the patch file and not any additional text.
Don't provide changes to test files.
Example prompt-response sequence:
# Example prompt 1
// FR0: App should print out 'Hello World!
package main

import "testing"

func TestGetMessage(t *testing.T) {
	want := "Hello World!"
	got := getMessage()
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

# Example response 1:
--- app/main.go
+++ app/main.go
@@ -0,0 +1,11 @@
+package main
+
+import "fmt"
+
+func main() {
+    fmt.Println(getText())
+}
+
+func getText() string {
+    return "Hello World!"
+}

# Example prompt 2
// FR1: Length of the printed out text should be 5 words
package main

import "testing"

func TestGetTextLength(t *testing.T) {
	want := 5
	got := len(strings.Split(getText(), " "))
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

# Example response 2:
--- app/main.go
+++ app/main.go
@@ -7,5 +7,5 @@
 }
 
 func getText() string {
-    return "Hello World!"
+    return "This text has 5 words!"
 }

`
)

const SystemPromptSWEBenchLike = `You will be provided with a partial codebase inside the app/ directory and an issue statement explaining what needs to be changed in that codebase.`

func PreparePatchPrompt(projectInfo project.ProjectInfo, i int) ([]byte, error) {
	requirement := projectInfo.Project.Requirements[i]
	codebase, err := project.LoadCodebase()
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
I need you to implement the required changes and only the required changes by generating a single patch file that can be applied directly using git apply. Please respond with a single patch file in the format shown above. Don't add any additional text or comments only the patch file contents.
Respond below:
`

	ret := []byte(fmt.Sprintf(prompt, requirement.Description, codebase))
	_ = os.WriteFile("data/prompt.txt", ret, 0644)

	return ret, nil
}

func PrepareWritePrompt(projectInfo project.ProjectInfo, i int) ([]byte, error) {
	requirement := projectInfo.Project.Requirements[i]
	codebase, err := project.LoadCodebase()
	if err != nil {
		return nil, fmt.Errorf("error loading codebase: %w", err)
	}
	blockMark := "```"
	prompt := `You will be provided with a partial codebase inside the app/ directory and an issue statement explaining what needs to be changed in that codebase. The changes made need to make all provided tests pass but you can't change or add any tests.
<issue>
%s
</issue>
<codebase>
%s
</codebase>
Here is an example of a generated list of files. It consists of code blocks annotated with the language. Comments specify the start and end of each file and its name.
<files>
%sgo
// start of file.go
package main

import "fmt"

func main() {
	fmt.Println(getNumber())
}
// end of file.go
// start of numbers.go
package main

func getNumber() int {
	return 7
}
// end of numbers.go
%s
</files>
I need you to implement changes needed to solve the issue and provide all files changed in the format shown above. Please respond with a list of non-test files in the format shown above. Don't add any additional text or comments only the file contents.
Respond below:
`

	ret := []byte(fmt.Sprintf(prompt, requirement.Description, codebase, blockMark, blockMark))
	_ = os.WriteFile("data/prompt.txt", ret, 0644)

	return ret, nil
}

// PrepareImagePrompt looks for attachments in the requirement and loads the first image,
// if no image is found it returns an error, if there are no attachments it returns nil
func PrepareImagePrompt(pInfo project.ProjectInfo, i int) ([]byte, error) {
	req := pInfo.Project.Requirements[i]
	if len(req.Attachments) > 0 {
		imgPath := fmt.Sprintf("templates/%s/%s", pInfo.Dir, req.Attachments[0])
		fmt.Printf("Loading image from %s ...\n", imgPath)
		imagePng, err := os.ReadFile(imgPath)
		if err != nil {
			return nil, fmt.Errorf("error reading image: %w", err)
		}
		return []byte(base64.StdEncoding.EncodeToString(imagePng)), nil
	}
	return nil, nil
}

type File struct {
	RelPath string
	Content []byte
}

// ParseWriteResponse parses the response from the user and extracts the file contents
//
// file contents are inside ``` blocks and annotated with // start of file.go and // end of file.go
// we need to extract the contents of the blocks and write them to files
// use regex ```((.|\n)*)``` to extract the blocks
func ParseWriteResponse(content []byte) ([]File, error) {
	re := regexp.MustCompile("```go\n((.|\n)*)```")
	matches := re.FindAllSubmatch(content, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no file contents found")
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple file contents found")
	}
	file := matches[0][1]
	re = regexp.MustCompile("(\\/\\/ start of )(.*\\..*)\n((.|\n)*?)(\\/\\/ end)")
	matches = re.FindAllSubmatch(file, -1)
	var files []File
	for _, match := range matches {
		newFile := File{
			RelPath: string(match[2]),
			Content: match[3],
		}
		files = append(files, newFile)
		// fmt.Printf("file: '%s'\ncontent: '\n%s'\n", match[2], match[3])
		// _ = os.WriteFile("data/"+newFile.Name, newFile.Content, 0644)
	}
	return files, nil
}
