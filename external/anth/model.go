package anth

import (
	"fmt"

	"github.com/Sph3ricalPeter/frbench/external"
)

type AnthModel string

const (
	Claude35Haiku  AnthModel = "claude-3-5-haiku-20241022"
	Claude3Haiku   AnthModel = "claude-3-haiku-20240307"
	Claude35Sonnet AnthModel = "claude-3-5-sonnet-20241022"
)

var (
	StrToAnthModel = map[string]AnthModel{
		"claude-3-5-haiku-20241022":  Claude35Haiku,
		"claude-3-haiku-20240307":    Claude3Haiku,
		"claude-3-5-sonnet-20241022": Claude35Sonnet,
	}
)

type AnthRole string

const (
	AnthRoleUser      AnthRole = "user"
	AnthRoleAssistant AnthRole = "assistant"
)

var (
	RoleToAnthRole = map[external.Role]AnthRole{
		external.RoleUser:  AnthRoleUser,
		external.RoleModel: AnthRoleAssistant,
	}
)

type AnthPromptData struct {
	Model AnthModel `json:"model"`
	Role  AnthRole  `json:"role"`
}

func mapPromptData(p external.SendPromptOpts) (*AnthPromptData, error) {
	if _, ok := RoleToAnthRole[p.Role]; !ok {
		return nil, fmt.Errorf("invalid role: %s", p.Role)
	}
	return &AnthPromptData{
		Role: RoleToAnthRole[p.Role],
	}, nil
}

type AnthMessage struct {
	Role    AnthRole             `json:"role"`
	Content []AnthMessageContent `json:"content"`
}

type AnthMessageContent struct {
	Type string `json:"type"`
	// content is either text or source, not both
	Text   string           `json:"text,omitempty"`
	Source *AnthImageSource `json:"source,omitempty"`
}

func NewTextContent(text string) AnthMessageContent {
	return AnthMessageContent{
		Type: "text",
		Text: text,
	}
}

func NewImageContent(mediaType, data string) AnthMessageContent {
	return AnthMessageContent{
		Type: "image",
		Source: &AnthImageSource{
			Type:      "base64",
			MediaType: mediaType,
			Data:      data,
		},
	}
}

type AnthImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

func NewMessage(role AnthRole, content []AnthMessageContent) AnthMessage {
	return AnthMessage{
		Role:    role,
		Content: content,
	}
}

type AnthRequest struct {
	Model     AnthModel     `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system,omitempty"`
	Messages  []AnthMessage `json:"messages"`
}

func NewRequest(model AnthModel, maxTokens int, system string, messages []AnthMessage) AnthRequest {
	return AnthRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  messages,
	}
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

type AnthErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
