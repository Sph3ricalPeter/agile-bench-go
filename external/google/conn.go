package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/internal"
	"github.com/Sph3ricalPeter/frbench/internal/common"
)

var (
	ApiKey = common.MustGetEnv("GOOGLE_API_KEY")
)

type GoogleConnector struct {
	model     GoogleModel
	sysPrompt string
	history   []GeminiMessage
	cache     *internal.JsonCache
}

func NewGoogleConnector(model GoogleModel, sysPrompt string) *GoogleConnector {
	return &GoogleConnector{
		model:     model,
		sysPrompt: sysPrompt,
		history:   make([]GeminiMessage, 0),
		cache:     internal.NewJsonCache("cache/google"),
	}
}

func (c *GoogleConnector) SendPrompt(pd external.SendPromptOpts) (*external.SendPromptResult, error) {
	gpd, err := mapToGoogleData(pd)
	if err != nil {
		return nil, fmt.Errorf("error mapping prompt data: %w", err)
	}

	promptMsg := NewGeminiMessage(gpd.Role, string(pd.Prompt))
	var msgs []GeminiMessage
	if pd.UseHistory {
		msgs = append(c.history, promptMsg)
	} else {
		msgs = []GeminiMessage{promptMsg}
	}
	reqPayload := NewGeminiRequest(c.sysPrompt, msgs)

	reqBody := bytes.NewBuffer([]byte{})
	json.NewEncoder(reqBody).Encode(reqPayload)

	// FIXME: testing only
	_ = os.WriteFile(fmt.Sprintf("data/gemini-req-%d.json", pd.Number), reqBody.Bytes(), 0644)

	cacheKey := internal.CreateCacheKey(pd.Prompt, pd.Number)
	if pd.UseCache {
		if respBytes, ok := c.cache.Get(cacheKey); ok {
			fmt.Println("Using cached response ...")
			c.history = append(c.history, promptMsg)
			return c.GetPromptResult(respBytes, true, &cacheKey)
		}
	}

	respBytes, err := sendRequest(reqBody, c.model)
	if err != nil {
		return nil, fmt.Errorf("error sending prompt: %w", err)
	}

	// FIXME: testing only
	_ = os.WriteFile(fmt.Sprintf("data/gemini-resp-%d.json", pd.Number), respBytes, 0644)

	if err := c.cache.Put(cacheKey, respBytes); err != nil {
		return nil, fmt.Errorf("error caching response: %w", err)
	}

	if pd.UseHistory {
		c.history = append(c.history, promptMsg)
	}

	return c.GetPromptResult(respBytes, false, &cacheKey)
}

func (c *GoogleConnector) GetPromptResult(resp []byte, isCached bool, cacheKey *string) (*external.SendPromptResult, error) {
	respData := GeminiResponse{}
	if err := json.Unmarshal(resp, &respData); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}
	if len(respData.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	c.history = append(c.history, respData.Candidates[0].Content)

	return &external.SendPromptResult{
		RespBytes: resp,
		Usage: external.ModelUsage{
			InputTokens:  respData.UsageMetadata.PromptTokenCount,
			OutputTokens: respData.UsageMetadata.CandidatesTokenCount,
		},
		Content:   respData.Candidates[0].Content.Parts[0].Text,
		CacheKey:  cacheKey,
		UsedCache: isCached,
	}, nil
}

func (c *GoogleConnector) InvalidateCachedPrompt(cacheKey string) error {
	return c.cache.Delete(cacheKey)
}

func (c *GoogleConnector) GetModelName() string {
	return string(c.model)
}

func sendRequest(reqBody *bytes.Buffer, model GoogleModel) ([]byte, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, ApiKey)
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending prompt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp GoogleErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("unexpected response: code=%d, msg=%s", errResp.Error.Code, errResp.Error.Message)
		}
		return nil, fmt.Errorf("unexpected response: %d", resp.StatusCode)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return respBytes, nil
}
