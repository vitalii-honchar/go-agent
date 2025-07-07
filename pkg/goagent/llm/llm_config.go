package llm

// LLMType represents the type of LLM provider
type LLMType string

const (
	// LLMTypeOpenAI represents the OpenAI LLM provider
	LLMTypeOpenAI LLMType = "openai"
)

// LLMConfig contains configuration for LLM providers
type LLMConfig struct {
	Type        LLMType `json:"type"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}
