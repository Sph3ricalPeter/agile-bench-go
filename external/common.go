package external

type Connector interface {
	// send a prompt to the model with a desired role,
	// optionally use cache, history or provide attachments to include in the prompt
	SendPrompt(pd SendPromptOpts) (*SendPromptResult, error)

	// unmarshal the response and return a common prompt result
	GetPromptResult(resp []byte, isCached bool, cacheKey *string) (*SendPromptResult, error)

	// cache a prompt response by key
	CacheResponse(cacheKey string, respByte []byte) error

	// invalidate a cached prompt by key
	InvalidateCachedPrompt(cacheKey string) error

	// get the model name string
	GetModelName() string
}

type Role string

const (
	RoleUser  Role = "user"
	RoleModel Role = "model"
)

type SendPromptOpts struct {
	Role       Role
	Prompt     []byte
	Image      []byte
	Number     int
	UseCache   bool
	UseHistory bool
}

func NewUserPromptOpts(prompt []byte, image []byte, number int, useCache bool) SendPromptOpts {
	return SendPromptOpts{
		Role:     RoleUser,
		Prompt:   prompt,
		Image:    image,
		Number:   number,
		UseCache: useCache,
	}
}

type SendPromptResult struct {
	RespBytes []byte
	Content   string
	Usage     ModelUsage
	UsedCache bool
	CacheKey  *string
}

type ModelUsage struct {
	InputTokens  int
	OutputTokens int
}

type PromptResultContent struct {
	Text string
}
