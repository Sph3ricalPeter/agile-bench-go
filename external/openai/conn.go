package openai

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
	ApiKey = common.MustGetEnv("OPENAI_API_KEY")
)

const (
	QuotaResetWindowSeconds = 60
	QuotaLimitWaitSeconds   = 5
)

var (
	ModelsCostMap = map[OpenAIModel]external.ModelCost{
		Gpt4oMini: {
			UsdMtokIn:  0.15,
			UsdMtokOut: 0.6,
		},
		Gpt4o: {
			UsdMtokIn:  2.5,
			UsdMtokOut: 10.0,
		},
		O1Mini: {
			UsdMtokIn:  3.0,
			UsdMtokOut: 12.0,
		},
		O1Preview: {
			UsdMtokIn:  15.0,
			UsdMtokOut: 60.0,
		},
	}
)

type OpenAIConnector struct {
	model     OpenAIModel
	sysPrompt string
	cache     *internal.JsonCache
}

func NewOpenAIConnector(model OpenAIModel, sysPrompt string) *OpenAIConnector {
	return &OpenAIConnector{
		model:     model,
		sysPrompt: sysPrompt,
		cache:     internal.NewJsonCache("cache/openai/" + string(model)),
	}
}

func (c *OpenAIConnector) SendPrompt(opts external.SendPromptOpts) (*external.SendPromptResult, error) {
	pd, err := mapPromptData(opts)
	if err != nil {
		return nil, fmt.Errorf("error mapping prompt data: %w", err)
	}

	content := []OpenAIMessageContent{
		NewTextContent(string(opts.Prompt)),
	}
	if opts.Image != nil {
		content = append(content, NewImageContent(string(opts.Image)))
	}

	reqPayload := NewRequest(c.model, external.MaxTokensPerPrompt, opts.Temp, NewOpenAIMessage(pd.Role, content))
	if c.model == O1Mini {
		// o1-mini only supports temp 1
		reqPayload.Temperature = 1.0
	}

	cacheKey := internal.CreateCacheKey(opts.Prompt, opts.Number)
	if opts.UseCache {
		if respBytes, ok := c.cache.Get(cacheKey); ok {
			fmt.Println(("Using cached response ..."))
			return c.GetPromptResult(respBytes, true, &cacheKey, 0)
		}
	}

	reqBody := bytes.NewBuffer([]byte{})
	common.CheckErr(json.NewEncoder(reqBody).Encode(reqPayload))
	// FIXME: testing only
	_ = os.WriteFile(fmt.Sprintf("data/openai-req-%d.json", opts.Number), reqBody.Bytes(), 0644)

	startTime := time.Now()
	respBytes, err := sendRequest(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error sending prompt: %w", err)
	}
	totalDuration := time.Since(startTime)

	return c.GetPromptResult(respBytes, false, &cacheKey, totalDuration)
}

func (c *OpenAIConnector) GetPromptResult(resp []byte, isCached bool, cacheKey *string, duration time.Duration) (*external.SendPromptResult, error) {
	respData := OpenAIResponse{}
	err := json.Unmarshal(resp, &respData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}
	if len(respData.Choices) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	return &external.SendPromptResult{
		RespBytes: resp,
		Usage: external.ModelUsage{
			InputTokens:  respData.Usage.PromptTokens,
			OutputTokens: respData.Usage.CompletionTokens,
		},
		Content:   respData.Choices[0].Message.Content,
		CacheKey:  cacheKey,
		UsedCache: isCached,
		Duration:  duration,
	}, nil
}

func (c *OpenAIConnector) CacheResponse(cacheKey string, respByte []byte) error {
	return c.cache.Put(cacheKey, respByte)
}

func (c *OpenAIConnector) InvalidateCachedPrompt(cacheKey string) error {
	return c.cache.Delete(cacheKey)
}

func (c *OpenAIConnector) GetModelName() string {
	return string(c.model)
}

func (c *OpenAIConnector) GetCost() external.ModelCost {
	return ModelsCostMap[c.model]
}

func sendRequest(reqBody *bytes.Buffer) ([]byte, error) {
loop:
	for i := 0; i < QuotaResetWindowSeconds/QuotaLimitWaitSeconds; i++ {
		reqBodyCopy := bytes.NewBuffer(reqBody.Bytes())
		client := &http.Client{}
		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", reqBodyCopy)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Add("Authorization", "Bearer "+ApiKey)
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %w", err)
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			fmt.Printf("Too many requests, waiting for %d seconds ...\n", QuotaLimitWaitSeconds)
			time.Sleep(QuotaLimitWaitSeconds * time.Second)
			continue loop
		case http.StatusOK:
			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response: %w", err)
			}
			return respBytes, nil
		default:
			var data any
			err := json.NewDecoder(resp.Body).Decode(&data)
			if err == nil {
				return nil, fmt.Errorf("unexpected response: status=%s, body=%v", resp.Status, data)
			}
			return nil, fmt.Errorf("unexpected response: status=%s, body=%s", resp.Status, data)
		}

	}

	return nil, fmt.Errorf("failed to fetch response within %ds retry window", QuotaResetWindowSeconds)
}
