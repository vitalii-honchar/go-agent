package llm

type LLMType string

const (
	LLMTypeOpenAI LLMType = "openai"
)

type LLMConfig struct {
	Type        LLMType `json:"type"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}
