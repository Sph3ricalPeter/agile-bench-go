package external

import "time"

const (
	MaxTokensPerPrompt = 2048
)

type Connector interface {
	// send a prompt to the model with a desired role,
	// optionally use cache, history or provide attachments to include in the prompt
	SendPrompt(pd SendPromptOpts) (*SendPromptResult, error)

	// unmarshal the response and return a common prompt result
	GetPromptResult(resp []byte, isCached bool, cacheKey *string, duration time.Duration) (*SendPromptResult, error)

	// cache a prompt response by key
	CacheResponse(cacheKey string, respByte []byte) error

	// invalidate a cached prompt by key
	InvalidateCachedPrompt(cacheKey string) error

	// get the model name string
	GetModelName() string

	// get the cost of the model in $/MTOK
	GetCost() ModelCost
}

type Role string

const (
	RoleUser   Role = "user"
	RoleModel  Role = "model"
	RoleSystem Role = "system"
)

type SendPromptOpts struct {
	Role       Role
	Prompt     []byte
	Image      []byte
	Number     int
	Temp       float64
	UseCache   bool
	UseHistory bool
}

func NewUserPromptOpts(prompt []byte, image []byte, number int, temp float64, useCache bool) SendPromptOpts {
	return SendPromptOpts{
		Role:     RoleUser,
		Prompt:   prompt,
		Image:    image,
		Number:   number,
		Temp:     temp,
		UseCache: useCache,
	}
}

type SendPromptResult struct {
	RespBytes []byte
	Content   string
	Usage     ModelUsage
	UsedCache bool
	CacheKey  *string
	Duration  time.Duration
}

type ModelUsage struct {
	InputTokens  int
	OutputTokens int
}

type PromptResultContent struct {
	Text string
}

type ModelCost struct {
	UsdMtokIn  float64
	UsdMtokOut float64
}

func NewModelCost(usdMtokIn, usdMtokOut float64) ModelCost {
	return ModelCost{
		UsdMtokIn:  usdMtokIn,
		UsdMtokOut: usdMtokOut,
	}
}

func MustCalcTotalCost(model string, usage ModelUsage, cost ModelCost) float64 {
	return cost.UsdMtokOut*float64(usage.OutputTokens)/1000000 + cost.UsdMtokIn*float64(usage.InputTokens)/1000000
}
