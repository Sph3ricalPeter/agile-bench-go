package anth

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
	ApiKey = common.MustGetEnv("ANTHROPIC_API_KEY")
)

const (
	QuotaResetWindowSeconds = 60
	QuotaLimitWaitSeconds   = 5
)

var (
	ModelsCostMap = map[AnthModel]external.ModelCost{
		Claude3Haiku: {
			UsdMtokIn:  0.25,
			UsdMtokOut: 1.25,
		},
		Claude35Sonnet: {
			UsdMtokIn:  3.0,
			UsdMtokOut: 15.0,
		},
	}
)

type AnthConnector struct {
	model     AnthModel
	sysPrompt string
	history   []AnthMessage
	cache     *internal.JsonCache
}

func NewAnthConnector(model AnthModel, sysPrompt string) *AnthConnector {
	return &AnthConnector{
		model:     model,
		sysPrompt: sysPrompt,
		history:   make([]AnthMessage, 0),
		cache:     internal.NewJsonCache("cache/anth/" + string(model)),
	}
}

// SendPrompt sends a prompt to the Anthropic API and returns the result.
// If the prompt is successfully sent, the response is cached.
func (c *AnthConnector) SendPrompt(opts external.SendPromptOpts) (*external.SendPromptResult, error) {
	// map to model specific prompt data
	pd, err := mapPromptData(opts)
	if err != nil {
		return nil, fmt.Errorf("error mapping prompt data: %w", err)
	}

	content := []AnthMessageContent{
		NewTextContent(string(opts.Prompt)),
	}
	if opts.Image != nil {
		content = append(content, NewImageContent("image/png", string(opts.Image)))
	}
	promptMsg := NewMessage(pd.Role, content)
	var msgs []AnthMessage
	if opts.UseHistory {
		msgs = append(c.history, promptMsg)
	} else {
		msgs = []AnthMessage{promptMsg}
	}
	reqPayload := NewRequest(c.model, external.MaxTokensPerPrompt, c.sysPrompt, opts.Temp, msgs)

	// TODO: caching and history should be common for all connectors ...
	// if history uses common prompt struct, this can be moved outside of the SendPrompt method
	// and the prompt is only added to history if response is OK, which means it can be done after the SendPrompt call
	cacheKey := internal.CreateCacheKey(opts.Prompt, opts.Number)
	if opts.UseCache {
		if respBytes, ok := c.cache.Get(cacheKey); ok {
			fmt.Println("Using cached response ...")
			c.history = append(c.history, promptMsg)
			return c.GetPromptResult(respBytes, true, &cacheKey, 0)
		}
	}

	reqBody := bytes.NewBuffer([]byte{})
	json.NewEncoder(reqBody).Encode(reqPayload)

	// FIXME: testing only
	_ = os.WriteFile(fmt.Sprintf("data/anth-req-%d.json", opts.Number), reqBody.Bytes(), 0644)

	startTime := time.Now()
	respBytes, err := sendRequest(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error sending prompt: %w", err)
	}
	totalDuration := time.Since(startTime)
	c.history = append(c.history, promptMsg)

	return c.GetPromptResult(respBytes, false, &cacheKey, totalDuration)
}

func (c *AnthConnector) GetPromptResult(resp []byte, isCached bool, cacheKey *string, duration time.Duration) (*external.SendPromptResult, error) {
	respData := AnthResponse{}
	err := json.Unmarshal(resp, &respData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}
	if len(respData.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// add response to the history
	// TODO: fix
	// c.history = append(c.history, AnthMessage{
	// 	Role:    AnthRoleAssistant,
	// 	Content: respData.Content[0].Text,
	// })

	return &external.SendPromptResult{
		RespBytes: resp,
		Usage: external.ModelUsage{
			InputTokens:  respData.Usage.InputTokens,
			OutputTokens: respData.Usage.OutputTokens,
		},
		Content:   respData.Content[0].Text,
		CacheKey:  cacheKey,
		UsedCache: isCached,
		Duration:  duration,
	}, nil
}

func (c *AnthConnector) CacheResponse(cacheKey string, respByte []byte) error {
	return c.cache.Put(cacheKey, respByte)
}

func (c *AnthConnector) InvalidateCachedPrompt(cacheKey string) error {
	return c.cache.Delete(cacheKey)
}

func (c *AnthConnector) GetModelName() string {
	return string(c.model)
}

func (c *AnthConnector) GetCost() external.ModelCost {
	return ModelsCostMap[c.model]
}

func sendRequest(reqBody *bytes.Buffer) ([]byte, error) {
loop:
	for i := 0; i < QuotaResetWindowSeconds/QuotaLimitWaitSeconds; i++ {
		reqBodyCopy := bytes.NewBuffer(reqBody.Bytes())
		client := &http.Client{}
		req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", reqBodyCopy)
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

		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			fmt.Printf("Too many requests, waiting for %d seconds ...\n", QuotaLimitWaitSeconds)
			time.Sleep(QuotaLimitWaitSeconds * time.Second)
			continue loop
		case http.StatusOK:
			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response body: %w", err)
			}
			return respBytes, nil
		default:
			var errResp AnthErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
				return nil, fmt.Errorf("unexpected response: code=%d, type=%s, msg=%s", resp.StatusCode, errResp.Error.Type, errResp.Error.Message)
			}
			return nil, fmt.Errorf("unexpected response: code=%d", resp.StatusCode)
		}
	}

	return nil, fmt.Errorf("failed to fetch response within %ds retry window", QuotaResetWindowSeconds)
}
