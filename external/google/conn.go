package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/internal"
	"github.com/Sph3ricalPeter/frbench/internal/common"
)

var (
	ApiKey = common.MustGetEnv("GOOGLE_API_KEY")
)

const (
	QuotaResetWindowSeconds = 60
	QuotaLimitWaitSeconds   = 5
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
		cache:     internal.NewJsonCache("cache/google/" + string(model)),
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

func (c *GoogleConnector) CacheResponse(cacheKey string, respByte []byte) error {
	return c.cache.Put(cacheKey, respByte)
}

func (c *GoogleConnector) InvalidateCachedPrompt(cacheKey string) error {
	return c.cache.Delete(cacheKey)
}

func (c *GoogleConnector) GetModelName() string {
	return string(c.model)
}

func sendRequest(reqBody *bytes.Buffer, model GoogleModel) ([]byte, error) {
loop:
	for i := 0; i < QuotaResetWindowSeconds/QuotaLimitWaitSeconds; i++ {
		reqBodyCopy := bytes.NewBuffer(reqBody.Bytes())
		client := &http.Client{}
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, ApiKey)
		req, err := http.NewRequest("POST", url, reqBodyCopy)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending prompt: %w", err)
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			fmt.Printf("Too many requests, waiting %d seconds ...\n", QuotaLimitWaitSeconds)
			time.Sleep(QuotaLimitWaitSeconds * time.Second)
			continue loop
		case http.StatusOK:
			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response body: %w", err)
			}
			return respBytes, nil
		default:
			var errResp GoogleErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
				return nil, fmt.Errorf("unexpected response: code=%d, msg=%s", errResp.Error.Code, errResp.Error.Message)
			}
			return nil, fmt.Errorf("unexpected response: %d", resp.StatusCode)
		}

	}

	return nil, fmt.Errorf("failed to fetch response within %ds retry window", QuotaResetWindowSeconds)
}
