package external

type Connector interface {
	SendPrompt(pd SendPromptData) (*SendPromptResult, error)
	GetPromptResult(resp []byte, cacheKey *string) (*SendPromptResult, error)
	InvalidateCachedPrompt(cacheKey string) error
	GetModelName() string
}

type Role string

const (
	RoleUser  Role = "user"
	RoleModel Role = "model"
)

type SendPromptData struct {
	Role     Role
	Prompt   []byte
	UseCache bool
	Number   int
}

type SendPromptResult struct {
	RespBytes []byte
	Content   string
	Usage     ModelUsage
	CacheKey  *string
}

type ModelUsage struct {
	InputTokens  int
	OutputTokens int
}

type PromptResultContent struct {
	Text string
}
