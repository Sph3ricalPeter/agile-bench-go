package external

type Connector interface {
	SendPrompt(pd SendPromptOpts) (*SendPromptResult, error)
	GetPromptResult(resp []byte, isCached bool, cacheKey *string) (*SendPromptResult, error)
	InvalidateCachedPrompt(cacheKey string) error
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
	Number     int
	UseCache   bool
	UseHistory bool
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
