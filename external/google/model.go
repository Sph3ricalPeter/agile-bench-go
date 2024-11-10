package google

import (
	"fmt"

	"github.com/Sph3ricalPeter/frbench/external"
)

type GoogleModel string

const (
	Gemini15Flash   GoogleModel = "gemini-1.5-flash-latest"
	Gemini15Flash8B GoogleModel = "gemini-1.5-flash-8b-001"
	Gemini10Pro     GoogleModel = "gemini-1.0-pro-001"
)

var (
	StrToGoogleModels = map[string]GoogleModel{
		"gemini-1.5-flash-latest": Gemini15Flash,
		"gemini-1.5-flash-8b-001": Gemini15Flash8B,
		"gemini-1.0-pro-001":      Gemini10Pro,
	}
)

type GeminiRole string

const (
	GeminiRoleUser  GeminiRole = "user"
	GeminiRoleModel GeminiRole = "model"
)

var (
	RoleToGeminiRole = map[external.Role]GeminiRole{
		external.RoleUser:  GeminiRoleUser,
		external.RoleModel: GeminiRoleModel,
	}
)

type GooglePromptData struct {
	Role GeminiRole
}

func mapToGoogleData(p external.SendPromptOpts) (*GooglePromptData, error) {
	if _, ok := RoleToGeminiRole[p.Role]; !ok {
		return nil, fmt.Errorf("unknown role %q", p.Role)
	}
	return &GooglePromptData{
		Role: RoleToGeminiRole[p.Role],
	}, nil
}

type GeminiRequest struct {
	SystemInstruction GeminiSystemInstruction `json:"system_instruction"`
	Contents          []GeminiMessage         `json:"contents"`
}

func NewGeminiRequest(sysInstr string, contents []GeminiMessage) GeminiRequest {
	return GeminiRequest{
		SystemInstruction: GeminiSystemInstruction{
			Parts: []GeminiMessagePart{
				{Text: sysInstr},
			},
		},
		Contents: contents,
	}
}

type GeminiSystemInstruction struct {
	Parts []GeminiMessagePart `json:"parts"`
}

type GeminiMessage struct {
	Role  GeminiRole          `json:"role"`
	Parts []GeminiMessagePart `json:"parts"`
}

func NewGeminiMessage(role GeminiRole, text string) GeminiMessage {
	return GeminiMessage{
		Role: role,
		Parts: []GeminiMessagePart{
			{Text: text},
		},
	}
}

type GeminiMessagePart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content GeminiMessage `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	}
}

type GoogleErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
