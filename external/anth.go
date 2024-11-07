package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Sph3ricalPeter/frbench/common"
)

var (
	ApiKeyAnthropic = common.ReqGetEnv("ANTHROPIC_API_KEY")
)

const (
	AnthropicUrl = "https://api.anthropic.com/v1/messages"
)

const (
	SystemPrompt = `You are a software engineer in an agile team tasked to create an issue-tracker application in Golang.
You will receive one functional requirement (FR) or change request (CR) for what you need to implement in each prompt and are expected to provide all required code and file changes (line additions, modifications and removals) in a patch file.
Along with the FR, you will receive the contents of a test file that will verify the changes you made. All previous tests should also pass.
Assume the working directory is app/ and it starts empty.
Assume your changes from previous prompts are already applied.
Your response should only include the patch file and not any additional text.
Make sure the changes don't result in any compilation errors.
Don't provide changes to test files.
Make sure line ranges of the patch file header include all changes provided in the patch file.
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
// CR1: Length of the printed out text should be 5 words
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

type AnthMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthMessagesPayload struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system,omitempty"`
	Messages  []AnthMessage `json:"messages"`
}

type AnthContent struct {
	Text string `json:"text"`
}

type AnthUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type AnthResponse struct {
	Id      string        `json:"id"`
	Role    string        `json:"role"`
	Content []AnthContent `json:"content"`
	Usage   AnthUsage     `json:"usage"`
}

func SendPrompt(payload AnthMessagesPayload) ([]byte, error) {
	data := bytes.NewBuffer([]byte{})
	json.NewEncoder(data).Encode(payload)

	client := &http.Client{}
	req, err := http.NewRequest("POST", AnthropicUrl, data)
	if err != nil {
		panic(err)
	}

	req.Header.Add("x-api-key", ApiKeyAnthropic)
	req.Header.Add("anthropic-version", "2023-06-01")
	req.Header.Add("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending prompt: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

func ParseResponse(data []byte) (*AnthResponse, error) {
	var response AnthResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}
	return &response, nil
}

func CreatePatchFile(response AnthResponse, fpath string) error {
	patchFile, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("error creating patch file: %w", err)
	}
	defer patchFile.Close()

	for _, content := range response.Content {
		if len(content.Text) == 0 || content.Text[len(content.Text)-1] != '\n' {
			content.Text += "\n"
		}

		_, err := patchFile.WriteString(content.Text)
		if err != nil {
			return fmt.Errorf("error writing to patch file: %w", err)
		}
	}

	return nil
}
