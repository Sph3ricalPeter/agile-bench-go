package openai

import (
	"fmt"

	"github.com/Sph3ricalPeter/frbench/external"
)

type OpenAIModel string

const (
	Gpt4oMini OpenAIModel = "gpt-4o-mini"
	Gpt4o     OpenAIModel = "gpt-4o"
	O1Mini    OpenAIModel = "o1-mini"
	O1Preview OpenAIModel = "o1-preview"
)

var (
	StrToOpenAIModel = map[string]OpenAIModel{
		"gpt-4o-mini": Gpt4oMini,
		"gpt-4o":      Gpt4o,
		"o1-mini":     O1Mini,
		"o1-preview":  O1Preview,
	}
)

type OpenAIRole string

const (
	OpenAIRoleUser      OpenAIRole = "user"
	OpenAIRoleAssistant OpenAIRole = "assistant"
	OpenAIRoleSystem    OpenAIRole = "system"
)

var (
	RoleToOpenAIRole = map[external.Role]OpenAIRole{
		external.RoleUser:   OpenAIRoleUser,
		external.RoleModel:  OpenAIRoleAssistant,
		external.RoleSystem: OpenAIRoleSystem,
	}
)

type OpenAIPromptData struct {
	Role OpenAIRole `json:"role"`
}

func mapPromptData(p external.SendPromptOpts) (*OpenAIPromptData, error) {
	if _, ok := RoleToOpenAIRole[p.Role]; !ok {
		return nil, fmt.Errorf("invalid role: %s", p.Role)
	}
	return &OpenAIPromptData{
		Role: RoleToOpenAIRole[p.Role],
	}, nil
}

type OpenAIMessage struct {
	Role    OpenAIRole             `json:"role"`
	Content []OpenAIMessageContent `json:"content"`
}

func NewOpenAIMessage(role OpenAIRole, content []OpenAIMessageContent) OpenAIMessage {
	return OpenAIMessage{
		Role:    role,
		Content: content,
	}
}

type OpenAIMessageContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageUrl *OpenAIImageUrl `json:"image_url,omitempty"`
}

type OpenAIImageUrl struct {
	Url string `json:"url"`
}

func NewTextContent(content string) OpenAIMessageContent {
	return OpenAIMessageContent{
		Type: "text",
		Text: content,
	}
}

func NewPngImageContent(datab64 string) OpenAIMessageContent {
	return OpenAIMessageContent{
		Type: "image_url",
		ImageUrl: &OpenAIImageUrl{
			Url: fmt.Sprintf("data:image/png;base64,%s", datab64),
		},
	}
}

type OpenAIRequest struct {
	Model       OpenAIModel     `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_completion_tokens"`
	Temperature float64         `json:"temperature"`
}

func NewRequest(model OpenAIModel, maxTokens int, temp float64, messages ...OpenAIMessage) OpenAIRequest {
	return OpenAIRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temp,
	}
}

type OpenAIResponse struct {
	Id      string                 `json:"id"`
	Choices []OpenAIResponseChoice `json:"choices"`
	Usage   OpenAIUsage            `json:"usage"`
}

type OpenAIResponseChoice struct {
	Message      OpenAIResponseMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

// OpenAI seemingly supports image content in response, but we are OK with text only
type OpenAIResponseMessage struct {
	Role    OpenAIRole `json:"role"`
	Content string     `json:"content"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}
