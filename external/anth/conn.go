package anth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Sph3ricalPeter/frbench/common"
	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/internal"
)

var (
	ApiKey = common.MustGetEnv("ANTHROPIC_API_KEY")
)

const (
	MaxTokens = 2048
)

type AnthConnector struct {
	model   AnthModel
	history []AnthMessage
	cache   *internal.JsonCache
}

func NewAnthConnector(model AnthModel) *AnthConnector {
	return &AnthConnector{
		model:   model,
		history: make([]AnthMessage, 0),
		cache:   internal.NewJsonCache("cache/anth"),
	}
}

// SendPrompt sends a prompt to the Anthropic API and returns the result.
// If the prompt is successfully sent, the response is cached.
func (c *AnthConnector) SendPrompt(pd external.SendPromptData) (*external.SendPromptResult, error) {
	// map to model specific prompt data
	apd, err := mapPromptData(pd)
	if err != nil {
		return nil, fmt.Errorf("error mapping prompt data: %w", err)
	}

	promptMsg := NewMessage(apd.Role, string(pd.Prompt))
	reqPayload := NewRequest(c.model, MaxTokens, common.SystemPrompt, append(c.history, promptMsg))

	reqBody := bytes.NewBuffer([]byte{})
	json.NewEncoder(reqBody).Encode(reqPayload)

	// FIXME: testing only
	_ = os.WriteFile(fmt.Sprintf("data/anth-req-%d.json", pd.Number), reqBody.Bytes(), 0644)

	cacheKey := common.CreateCacheKey(pd.Prompt, pd.Number)
	if pd.UseCache {
		if respBytes, ok := c.cache.Get(cacheKey); ok {
			fmt.Println("Using cached response ...")
			c.history = append(c.history, promptMsg)
			return c.GetPromptResult(respBytes, &cacheKey)
		}
	}

	respBytes, err := sendRequest(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error sending prompt: %w", err)
	}

	err = c.cache.Put(cacheKey, respBytes)
	if err != nil {
		return nil, fmt.Errorf("error caching prompt: %w", err)
	}

	// add request to the history
	c.history = append(c.history, promptMsg)

	return c.GetPromptResult(respBytes, &cacheKey)
}

func (c *AnthConnector) GetPromptResult(resp []byte, cacheKey *string) (*external.SendPromptResult, error) {
	respData := AnthResponse{}
	err := json.Unmarshal(resp, &respData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}
	if len(respData.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// add response to the history
	c.history = append(c.history, AnthMessage{
		Role:    AnthRoleAssistant,
		Content: respData.Content[0].Text,
	})

	return &external.SendPromptResult{
		RespBytes: resp,
		Usage: external.ModelUsage{
			InputTokens:  respData.Usage.InputTokens,
			OutputTokens: respData.Usage.OutputTokens,
		},
		Content:  respData.Content[0].Text,
		CacheKey: cacheKey,
	}, nil
}

func (c *AnthConnector) InvalidateCachedPrompt(cacheKey string) error {
	return c.cache.Delete(cacheKey)
}

func (c *AnthConnector) GetModelName() string {
	return string(c.model)
}

func sendRequest(reqBody *bytes.Buffer) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("x-api-key", ApiKey)
	req.Header.Add("anthropic-version", "2023-06-01")
	req.Header.Add("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending prompt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp AnthErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("unexpected response: code=%d, type=%s, msg=%s", resp.StatusCode, errResp.Error.Type, errResp.Error.Message)
		}
		return nil, fmt.Errorf("unexpected response: code=%d", resp.StatusCode)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return respBytes, nil
}
