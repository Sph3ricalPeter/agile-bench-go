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
You will receive one functional requirement (FR) for what you need to implement in each prompt and are expected to provide all required code and file changes in a patch file. 
Along with the FR, you will receive the contents of a test file that will verify the implementation of the FR.
Assume the working directory is app/ and it starts empty. Your response should only include the patch file and not any additional text.
Example prompt:
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
Desired response:
--- app/main.go
+++ app/main.go
@@ -0,0 +1,11 @@
+package main
+
+import "fmt"
+
+func main() {
+   fmt.Println(getMessage())
+}
+
+func getMessage() string {
+   return "Hello World!"
+}
`
)

type AnthContent struct {
	Text string `json:"text"`
}

type AnthResponse struct {
	Id      string        `json:"id"`
	Role    string        `json:"role"`
	Content []AnthContent `json:"content"`
}

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

func SendPrompt() ([]byte, error) {
	payload := AnthMessagesPayload{
		Model:     "claude-3-haiku-20240307",
		MaxTokens: 1024,
		System:    SystemPrompt,
		Messages: []AnthMessage{
			{
				Role: "user",
				Content: `
package main

import "testing"

// FR1: App should print out 'This will be an issue tracker!
func TestGetMessage(t *testing.T) {
	want := "This will be an issue tracker!"
	got := getMessage()
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}
`,
			},
		},
	}

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
		patchFile.WriteString(content.Text)
	}

	return nil
}
